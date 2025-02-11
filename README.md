# ClickUp Tasks Report Tool

## Overview

This is a Go-based CLI tool that fetches and organizes tasks from ClickUp using their API. The tool retrieves tasks from different lists, sorts them by priority and due date, and presents them in a structured table format in the terminal.

## Features

- Fetches tasks from ClickUp lists using the API.
- Supports authentication via API key.
- Categorizes tasks by status: **To Do**, **In Progress**, and **Completed**.
- Sorts tasks based on priority and due date.
- Displays tasks in a structured tabular format.
- Provides environment variable support for API credentials.

## Prerequisites

- Go (>=1.23)
- A ClickUp API Key
- ClickUp Space ID

## Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/your-repo/clickup-task-reporter.git
   cd clickup-task-reporter
   ```

2. Build the executable:
   ```sh
   go build -o clickup-tasks
   ```

## Usage

### Running the tool with flags

```sh
./clickup-tasks -api-key=YOUR_CLICKUP_API_KEY -space-id=YOUR_SPACE_ID
```

Or using shorthand flags:

```sh
./clickup-tasks -k=YOUR_CLICKUP_API_KEY -s=YOUR_SPACE_ID
```

### Running the tool with environment variables

You can set environment variables instead of using command-line flags:

```sh
export CLICKUP_API_KEY=YOUR_CLICKUP_API_KEY
export CLICKUP_SPACE_ID=YOUR_SPACE_ID
./clickup-tasks
```

## Output

The tool will categorize tasks into three sections:

1. **To Do Tasks** - Tasks that are yet to be worked on.
2. **In Progress Tasks** - Tasks that are currently being worked on.
3. **Completed Tasks** - Tasks that have been marked as finished.

### Example Output

```
========================================
ClickUp Tasks Report
========================================
Fetching Tasks:
----------------------------------------
Fetching tasks from list: Project Alpha
Fetching tasks from list: Feature Requests

========================================
Task Summary:
------------
Completed Tasks: 12
To Do Tasks: 8
In Progress Tasks: 5
----------------------------------------

Tasks by Status:
To Do Tasks:
Task Name        List              Due Date   Priority  
------------------------------------------------------
Task 1          Project Alpha     2024-02-15 High      
Task 2          Feature Requests  2024-02-18 Normal    
------------------------------------------------------
```

## Configuration & Customization

### API Credentials
- API Key and Space ID can be passed as flags or set as environment variables.

### Task Sorting
- Tasks are sorted by **priority** first, then by **due date**.

### ANSI Colors
- Red: Urgent tasks
- Yellow: High priority tasks
- Blue: Normal priority tasks
- No color: Low priority or undefined tasks

## Error Handling
If credentials are missing or the API request fails, the tool will display an appropriate error message.

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request if you have improvements or bug fixes.

