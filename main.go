package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	apiKey  string
	spaceID string
	// Statuses that represent completion
	completedStatuses = map[string]bool{
		"complete":  true,
		"completed": true,
		"done":      true,
		"closed":    true,
		"finished":  true,
	}
	// ANSI color codes
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGreen  = "\033[32m"
	colorReset  = "\033[0m"
)

type Task struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Priority Priority `json:"priority"`
	DueDate  string   `json:"due_date"`
	Status   struct {
		Status string `json:"status"`
	} `json:"status"`
	List struct {
		Name string `json:"name"`
	} `json:"list"`
	DateClosed string `json:"date_closed"`
}

type Priority struct {
	Priority string `json:"priority"`
	Color    string `json:"color"`
}

type TasksResponse struct {
	Tasks []Task `json:"tasks"`
}

type List struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type FolderResponse struct {
	Folders []struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Lists []List `json:"lists"`
	} `json:"folders"`
}

type ListsResponse struct {
	Lists []List `json:"lists"`
}

func getPriorityLevel(priority Priority) int {
	switch strings.ToLower(priority.Priority) {
	case "urgent", "1":
		return 1
	case "high", "2":
		return 2
	case "normal", "medium", "3":
		return 3
	case "low", "4":
		return 4
	default:
		return 5
	}
}

func getPriorityColor(priority Priority) string {
	level := getPriorityLevel(priority)
	switch level {
	case 1:
		return colorRed
	case 2:
		return colorYellow
	case 3:
		return colorBlue
	default:
		return colorReset
	}
}

func parseDueDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	msec, err := strconv.ParseInt(dateStr, 10, 64)
	if err != nil {
		fmt.Println(err)
		return time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	}

	return time.Unix(msec/1000, (msec%1000)*1000000).UTC()
}

func sortTasks(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		iDue := parseDueDate(tasks[i].DueDate)
		jDue := parseDueDate(tasks[j].DueDate)

		if !iDue.Equal(jDue) {
			return iDue.Before(jDue)
		}

		iPriority := getPriorityLevel(tasks[i].Priority)
		jPriority := getPriorityLevel(tasks[j].Priority)
		return iPriority < jPriority
	})
}

func isCompletedStatus(status string) bool {
	statusLower := strings.ToLower(status)
	return completedStatuses[statusLower]
}

func isInProgressStatus(status string) bool {
	statusLower := strings.ToLower(status)
	return strings.Contains(statusLower, "progress") || strings.Contains(statusLower, "in progress")
}

func loadCredentials() error {
	flag.StringVar(&apiKey, "api-key", "", "ClickUp API Key")
	flag.StringVar(&spaceID, "space-id", "", "ClickUp Space ID")

	// Also allow short versions of flags
	flag.StringVar(&apiKey, "k", "", "ClickUp API Key (shorthand)")
	flag.StringVar(&spaceID, "s", "", "ClickUp Space ID (shorthand)")

	flag.Parse()

	// Check environment variables if flags are not provided
	if apiKey == "" {
		apiKey = os.Getenv("CLICKUP_API_KEY")
	}
	if spaceID == "" {
		spaceID = os.Getenv("CLICKUP_SPACE_ID")
	}

	// Validate credentials
	if apiKey == "" {
		return fmt.Errorf("ClickUp API Key is required. Provide it via -api-key flag or CLICKUP_API_KEY environment variable")
	}
	if spaceID == "" {
		return fmt.Errorf("ClickUp Space ID is required. Provide it via -space-id flag or CLICKUP_SPACE_ID environment variable")
	}

	return nil
}

func makeRequest(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	return client.Do(req)
}

func getAllLists() ([]List, error) {
	var allLists []List

	// Get folders and their lists
	folderEndpoint := fmt.Sprintf("https://api.clickup.com/api/v2/space/%s/folder", spaceID)
	resp, err := makeRequest(folderEndpoint)
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch folders: %s", resp.Status)
	}

	var folderResp FolderResponse
	if err := json.Unmarshal(body, &folderResp); err != nil {
		return nil, fmt.Errorf("error parsing folders response: %v", err)
	}

	// Add lists from folders
	for _, folder := range folderResp.Folders {
		allLists = append(allLists, folder.Lists...)
	}

	// Get folderless lists
	listEndpoint := fmt.Sprintf("https://api.clickup.com/api/v2/space/%s/list?archived=false", spaceID)
	resp, err = makeRequest(listEndpoint)
	if err != nil {
		return nil, err
	}
	body, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch folderless lists: %s", resp.Status)
	}

	var listsResp ListsResponse
	if err := json.Unmarshal(body, &listsResp); err != nil {
		return nil, fmt.Errorf("error parsing lists response: %v", err)
	}

	allLists = append(allLists, listsResp.Lists...)

	return allLists, nil
}

func getTasksForList(listID string) ([]Task, error) {
	endpoint := fmt.Sprintf("https://api.clickup.com/api/v2/list/%s/task?subtasks=true&include_closed=true", listID)
	resp, err := makeRequest(endpoint)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch tasks: %s", resp.Status)
	}

	var tasksResp TasksResponse
	if err := json.Unmarshal(body, &tasksResp); err != nil {
		return nil, fmt.Errorf("error parsing tasks response: %v", err)
	}

	return tasksResp.Tasks, nil
}

func formatTableRow(task Task) string {
	listName := task.List.Name
	if listName == "" {
		listName = "No List"
	}

	dueDate := parseDueDate(task.DueDate)
	dueDateStr := "No due date"
	if !dueDate.Equal(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)) {
		dueDateStr = dueDate.Format("2006-01-02")
	}

	priorityColor := getPriorityColor(task.Priority)
	priorityStr := task.Priority.Priority
	if priorityStr == "" {
		priorityStr = "None"
	}

	return fmt.Sprintf("%s%s\t%s\t%s\t%s%s",
		priorityColor,
		task.Name,
		listName,
		dueDateStr,
		priorityStr,
		colorReset)
}

func printDivider(char string, length int) {
	fmt.Println(strings.Repeat(char, length))
}

func printTaskTable(w *tabwriter.Writer, tasks []Task, title string) {
	dividerWidth := 80
	printDivider("=", dividerWidth)
	fmt.Printf("\n%s:\n", title)
	printDivider("-", len(title)+1)

	fmt.Fprintln(w, "Task Name\tList\tDue Date\tPriority\t")
	fmt.Fprintln(w, strings.Repeat("-", 20)+"\t"+strings.Repeat("-", 15)+"\t"+strings.Repeat("-", 10)+"\t"+strings.Repeat("-", 8)+"\t")

	for _, task := range tasks {
		fmt.Fprintln(w, formatTableRow(task))
	}

	w.Flush()
	printDivider("-", dividerWidth)
	fmt.Println()
}

func printTasksByList(tasks []Task, title string) {
	dividerWidth := 80
	printDivider("=", dividerWidth)
	fmt.Printf("\n%s by List:\n", title)
	printDivider("-", len(title)+9)

	// Group tasks by list
	tasksByList := make(map[string][]Task)
	for _, task := range tasks {
		listName := task.List.Name
		if listName == "" {
			listName = "No List"
		}
		tasksByList[listName] = append(tasksByList[listName], task)
	}

	// Get sorted list names
	var listNames []string
	for listName := range tasksByList {
		listNames = append(listNames, listName)
	}
	sort.Strings(listNames)

	// Create a new tabwriter for each list
	for _, listName := range listNames {
		tasks := tasksByList[listName]
		sortTasks(tasks) // Sort tasks within each list

		fmt.Printf("\nList: %s (%d tasks)\n", listName, len(tasks))
		printDivider("~", 40)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "Task Name\tDue Date\tPriority\t")
		fmt.Fprintln(w, strings.Repeat("-", 20)+"\t"+strings.Repeat("-", 10)+"\t"+strings.Repeat("-", 8)+"\t")

		for _, task := range tasks {
			priorityColor := getPriorityColor(task.Priority)
			priorityStr := task.Priority.Priority
			if priorityStr == "" {
				priorityStr = "None"
			}

			dueDate := parseDueDate(task.DueDate)
			dueDateStr := "No due date"
			if !dueDate.Equal(time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)) {
				dueDateStr = dueDate.Format("2006-01-02")
			}

			fmt.Fprintf(w, "%s%s\t%s\t%s%s\n",
				priorityColor,
				task.Name,
				dueDateStr,
				priorityStr,
				colorReset)
		}
		w.Flush()
		fmt.Println()
	}
	printDivider("-", dividerWidth)
	fmt.Println()
}

func isCompletedThisWeek(task Task) bool {
	if task.DateClosed == "" {
		return false
	}

	msec, err := strconv.ParseInt(task.DateClosed, 10, 64)
	if err != nil {
		return false
	}

	closeDate := time.Unix(msec/1000, (msec%1000)*1000000).UTC()
	now := time.Now().UTC()

	// Get the start of the current week (Sunday)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, time.UTC)

	// Get the end of the current week (Saturday)
	weekEnd := weekStart.AddDate(0, 0, 7)

	return closeDate.After(weekStart) && closeDate.Before(weekEnd)
}

func main() {
	dividerWidth := 80
	printDivider("=", dividerWidth)
	fmt.Println("ClickUp Tasks Report")
	printDivider("=", dividerWidth)
	fmt.Println()

	if err := loadCredentials(); err != nil {
		fmt.Printf("Configuration error: %v\n", err)
		fmt.Println("\nUsage:")
		fmt.Println("  clickup-tasks -api-key=YOUR_API_KEY -space-id=YOUR_SPACE_ID")
		fmt.Println("  clickup-tasks -k=YOUR_API_KEY -s=YOUR_SPACE_ID")
		fmt.Println("\nYou can also set environment variables:")
		fmt.Println("  CLICKUP_API_KEY=YOUR_API_KEY")
		fmt.Println("  CLICKUP_SPACE_ID=YOUR_SPACE_ID")
		os.Exit(1)
	}

	lists, err := getAllLists()
	if err != nil {
		fmt.Printf("Error getting lists: %v\n", err)
		os.Exit(1)
	}

	var completedTasks []Task
	var todoTasks []Task
	var inProgressTasks []Task

	fmt.Println("Fetching Tasks:")
	printDivider("-", dividerWidth)

	for _, list := range lists {
		fmt.Printf("Fetching tasks from list: %s\n", list.Name)
		tasks, err := getTasksForList(list.ID)
		if err != nil {
			fmt.Printf("Error getting tasks for list %s: %v\n", list.Name, err)
			continue
		}

		for _, task := range tasks {
			status := task.Status.Status
			if status == "" {
				status = "to do"
			}

			if isCompletedStatus(status) {
				completedTasks = append(completedTasks, task)
			} else if isInProgressStatus(status) {
				inProgressTasks = append(inProgressTasks, task)
			} else {
				todoTasks = append(todoTasks, task)
			}
		}
	}

	// Sort all task slices
	sortTasks(todoTasks)
	sortTasks(inProgressTasks)

	// Print summary
	printDivider("=", dividerWidth)
	fmt.Println("\nTask Summary:")
	printDivider("-", 12)
	fmt.Printf("Completed Tasks: %d\n", len(completedTasks))
	fmt.Printf("To Do Tasks: %d\n", len(todoTasks))
	fmt.Printf("In Progress Tasks: %d\n", len(inProgressTasks))
	printDivider("-", dividerWidth)
	fmt.Println()

	// Create a tabwriter for status-based tables
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	// Print status-based tables
	fmt.Println("Tasks by Status:")
	if len(todoTasks) > 0 {
		printTaskTable(w, todoTasks, "To Do Tasks")
	}

	if len(inProgressTasks) > 0 {
		printTaskTable(w, inProgressTasks, "In Progress Tasks")
	}

	// Print list-based tables
	fmt.Println("Tasks by List:")
	if len(todoTasks) > 0 {
		printTasksByList(todoTasks, "To Do Tasks")
	}

	if len(inProgressTasks) > 0 {
		printTasksByList(inProgressTasks, "In Progress Tasks")
	}

	completedThisWeek := 0
	for _, task := range completedTasks {
		if isCompletedThisWeek(task) {
			completedThisWeek++
		}
	}

	printDivider("=", dividerWidth)
	fmt.Println("\nWeekly Completion Summary:")
	printDivider("-", 23)
	fmt.Printf("%sTasks completed this week: %d%s\n", colorGreen, completedThisWeek, colorReset)
	printDivider("-", dividerWidth)
	fmt.Println()

	printDivider("=", dividerWidth)
	fmt.Println("End of Report")
	printDivider("=", dividerWidth)
}
