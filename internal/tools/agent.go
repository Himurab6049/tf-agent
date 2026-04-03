package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// SubAgentRunner is a function that runs a sub-agent and returns its output.
// Injected at wiring time to avoid an import cycle (agent → tools → agent).
type SubAgentRunner func(ctx context.Context, prompt, role string, timeoutSecs int) (string, error)

// AgentTool spawns an isolated sub-agent to handle a focused task.
type AgentTool struct {
	runner SubAgentRunner
}

func NewAgentTool(runner SubAgentRunner) *AgentTool {
	return &AgentTool{runner: runner}
}

func (t *AgentTool) Name() string { return "Agent" }

func (t *AgentTool) Description() string {
	return "Spawn an isolated sub-agent to handle a focused task. The sub-agent has its own context and a role-appropriate tool subset."
}

func (t *AgentTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"required": ["prompt"],
		"properties": {
			"prompt": {
				"type": "string",
				"description": "Task for the sub-agent to complete"
			},
			"role": {
				"type": "string",
				"enum": ["reviewer", "coder", "tester", "security-auditor"],
				"description": "Role preset controlling which tools are available"
			},
			"timeout_seconds": {
				"type": "integer",
				"description": "Max seconds (default 120)"
			}
		}
	}`)
}

func (t *AgentTool) IsReadOnly() bool                     { return false }
func (t *AgentTool) IsDestructive(_ json.RawMessage) bool { return false }

func (t *AgentTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Prompt         string `json:"prompt"`
		Role           string `json:"role"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("Agent: invalid input: %w", err)
	}
	if args.Prompt == "" {
		return "", fmt.Errorf("Agent: prompt is required")
	}
	if args.TimeoutSeconds <= 0 {
		args.TimeoutSeconds = 120
	}
	return t.runner(ctx, args.Prompt, args.Role, args.TimeoutSeconds)
}
