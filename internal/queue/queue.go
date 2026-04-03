package queue

import "context"

// Item is a task queued for execution.
type Item struct {
	TaskID     string
	UserID     string
	QueueName  string // named queue to route this task; empty means "default"
	InputType  string // prompt | jira
	InputText  string
	OutputType string // pr | files | print
	OutputDir  string

	// Per-request credentials — never persisted.
	GitHubToken     string
	RepoURL         string
	AtlassianToken  string
	AtlassianDomain string
	AtlassianEmail  string
}

// Queue is the task submission interface.
// Phase 1: in-memory channel.
// Phase 2+: swap to Redis without changing callers.
type Queue interface {
	Push(ctx context.Context, item Item) error
	Pop(ctx context.Context) (Item, error)
	Len() int
}
