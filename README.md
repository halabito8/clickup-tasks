# ClickUp Tasks Reporter

## Overview
This Go application fetches and displays tasks from ClickUp, organizing them based on status and priority. It retrieves tasks from multiple lists, categorizes them into "To Do", "In Progress", and "Completed", and sorts them by priority and due date.

## Features
- Fetches all lists and tasks from a ClickUp workspace.
- Categorizes tasks based on their status (To Do, In Progress, Completed).
- Sorts tasks by priority and due date.
- Displays tasks in a structured table format.

## Prerequisites
- [Go](https://go.dev/) (version 1.23 or later recommended)
- A ClickUp API Key
- A ClickUp Space ID

## Installation
1. Clone the repository:
   ```sh
   git clone https://github.com/your-repository/clickup-tasks-reporter.git
   cd clickup-tasks-reporter
   ```
2. Install dependencies:
   ```sh
   go mod tidy
   ```

## Configuration
Create a `.env` file in the root directory and add your ClickUp API key and Space ID:

```
CLICKUP_API_KEY=your_clickup_api_key
CLICKUP_SPACE_ID=your_space_id
```

## Running the Application
Run the following command to execute the program:

```sh
 go run main.go
```

## How It Works
1. Loads API credentials from `.env`.
2. Fetches lists and tasks from ClickUp.
3. Categorizes tasks based on their status:
   - **Completed**: Tasks marked as complete.
   - **In Progress**: Tasks containing "progress" in their status.
   - **To Do**: All other tasks.
4. Sorts tasks by priority (Urgent → Low) and due date (Soonest → Latest).
5. Displays categorized tasks in a structured table format.

## Task Sorting Criteria
- **Priority (from highest to lowest)**:
  - Urgent
  - High
  - Normal
  - Low
- **Due Date**:
  - Tasks with an earlier due date appear first.
  - Tasks without a due date are listed last.

## Example Output
```
========================================
ClickUp Tasks Report
========================================
Fetching Tasks:
------------------------------
Fetching tasks from list: Project Alpha
Fetching tasks from list: Sprint Backlog

Task Summary:
----------------------
Completed Tasks: 5
To Do Tasks: 12
In Progress Tasks: 3
----------------------

Tasks by Status:
==========================
To Do Tasks:
--------------------------
Task Name        List Name      Due Date    Priority  
----------------------------------------------
Fix UI Bug       Project Alpha  2024-02-15  High      
Add API Testing  Sprint Backlog 2024-02-20  Normal    
--------------------------

Tasks by List:
==========================
Sprint Backlog (3 tasks)
~~~~~~~~~~~~~~~~~~~~~~~~
Task Name        Due Date    Priority  
----------------------------------------------
Add API Testing  2024-02-20  Normal    
----------------------------------------------
```

## Error Handling
- If the API key or space ID is missing, the program exits with an error.
- If an API request fails, the error is displayed, and the program continues with available data.

## Dependencies
- `github.com/joho/godotenv` (For loading environment variables)

## License
This project is licensed under the MIT License.
