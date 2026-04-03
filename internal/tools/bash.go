package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// BashTool executes shell commands with safety guardrails.
type BashTool struct{}

func (t *BashTool) Name() string        { return "bash" }
func (t *BashTool) IsReadOnly() bool    { return false }
func (t *BashTool) IsDestructive(_ json.RawMessage) bool { return true }

func (t *BashTool) Description() string {
	return "Execute a shell command. Timeout defaults to 120 seconds. Use for running scripts, installing packages, checking git status, etc."
}

func (t *BashTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {
				"type": "string",
				"description": "The shell command to execute"
			},
			"timeout": {
				"type": "integer",
				"description": "Timeout in seconds (default 120, max 600)"
			}
		},
		"required": ["command"]
	}`)
}

var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// blockedPatterns are shell patterns that must never be executed.
var blockedPatterns = []string{
	"rm -rf /",
	":(){ :|:& };:",
	"mkfs",
	"dd if=/dev/zero",
}

func (t *BashTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Command string `json:"command"`
		Timeout int    `json:"timeout"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("bash: invalid input: %w", err)
	}
	if args.Command == "" {
		return "", fmt.Errorf("bash: command is required")
	}

	// Safety check.
	for _, blocked := range blockedPatterns {
		if strings.Contains(args.Command, blocked) {
			return "", fmt.Errorf("bash: command blocked for safety: contains %q", blocked)
		}
	}

	timeout := args.Timeout
	if timeout <= 0 {
		timeout = 120
	}
	if timeout > 600 {
		timeout = 600
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", args.Command)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := stdout.String()
	if stderr.Len() > 0 {
		if out != "" {
			out += "\n"
		}
		out += stderr.String()
	}

	// Strip ANSI codes.
	out = ansiEscape.ReplaceAllString(out, "")

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return out, fmt.Errorf("bash: command timed out after %ds", timeout)
		}
		// Return output + error message (exit code errors are informative).
		if out == "" {
			out = err.Error()
		}
	}

	if len(out) > 100_000 {
		out = out[:100_000] + "\n... (truncated)"
	}
	return out, nil
}
