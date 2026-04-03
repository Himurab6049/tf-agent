package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Task represents a tracked task item.
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"` // "pending" | "in_progress" | "done"
	CreatedAt time.Time `json:"created_at"`
}

var (
	taskMu    sync.Mutex
	taskStore = make(map[string]*Task)
	taskSeq   int
)

// TaskTool provides create/update/list operations for tasks.
type TaskTool struct{}

func (t *TaskTool) Name() string                         { return "task" }
func (t *TaskTool) IsReadOnly() bool                     { return false }
func (t *TaskTool) IsDestructive(_ json.RawMessage) bool { return false }

func (t *TaskTool) Description() string {
	return "Manage a task list. Actions: create, update, list."
}

func (t *TaskTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"action": {
				"type": "string",
				"enum": ["create", "update", "list"],
				"description": "Operation to perform"
			},
			"id": {
				"type": "string",
				"description": "Task ID (required for update)"
			},
			"title": {
				"type": "string",
				"description": "Task title (required for create)"
			},
			"status": {
				"type": "string",
				"enum": ["pending", "in_progress", "done"],
				"description": "Task status (required for update)"
			}
		},
		"required": ["action"]
	}`)
}

func (t *TaskTool) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Action string `json:"action"`
		ID     string `json:"id"`
		Title  string `json:"title"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("task: invalid input: %w", err)
	}

	taskMu.Lock()
	defer taskMu.Unlock()

	switch args.Action {
	case "create":
		if args.Title == "" {
			return "", fmt.Errorf("task create: title is required")
		}
		taskSeq++
		id := fmt.Sprintf("task_%d", taskSeq)
		taskStore[id] = &Task{
			ID:        id,
			Title:     args.Title,
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		return fmt.Sprintf("Created task %s: %s", id, args.Title), nil

	case "update":
		if args.ID == "" {
			return "", fmt.Errorf("task update: id is required")
		}
		task, ok := taskStore[args.ID]
		if !ok {
			return "", fmt.Errorf("task update: task %s not found", args.ID)
		}
		if args.Status != "" {
			task.Status = args.Status
		}
		if args.Title != "" {
			task.Title = args.Title
		}
		return fmt.Sprintf("Updated task %s: status=%s", task.ID, task.Status), nil

	case "list":
		if len(taskStore) == 0 {
			return "No tasks", nil
		}
		var sb strings.Builder
		for _, task := range taskStore {
			fmt.Fprintf(&sb, "[%s] %s — %s\n", task.Status, task.ID, task.Title)
		}
		return sb.String(), nil

	default:
		return "", fmt.Errorf("task: unknown action %q", args.Action)
	}
}
