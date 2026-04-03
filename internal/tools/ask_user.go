package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tf-agent/tf-agent/internal/taskctx"
)

// AskUserTool pauses the agent and asks the user a clarifying question.
// The task runner injects the blocking implementation via taskctx.WithAskUser.
type AskUserTool struct{}

func (t *AskUserTool) Name() string        { return "ask_user" }
func (t *AskUserTool) IsReadOnly() bool    { return true }
func (t *AskUserTool) Description() string {
	return "Pause execution and ask the user a clarifying question. Use this when you need information that cannot be inferred from the task description or codebase. Returns the user's answer."
}
func (t *AskUserTool) IsDestructive(_ json.RawMessage) bool { return false }

func (t *AskUserTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"question": {
				"type": "string",
				"description": "The question to ask the user"
			}
		},
		"required": ["question"]
	}`)
}

func (t *AskUserTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Question string `json:"question"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("ask_user: invalid input: %w", err)
	}
	if args.Question == "" {
		return "", fmt.Errorf("ask_user: question is required")
	}

	answer, err := taskctx.AskUser(ctx, args.Question)
	if err != nil {
		return "", err
	}
	if answer == "" {
		return "User did not respond.", nil
	}
	return answer, nil
}
