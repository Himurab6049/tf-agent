package hooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/tf-agent/tf-agent/internal/config"
)

// Runner executes hooks around tool calls.
type Runner struct {
	hooks []Hook
}

// NewRunner builds a Runner from config.
func NewRunner(cfg *config.HooksConfig) *Runner {
	var hooks []Hook
	for _, h := range cfg.PreToolUse {
		hooks = append(hooks, Hook{Type: PreToolUse, Tool: h.Tool, Command: h.Command})
	}
	for _, h := range cfg.PostToolUse {
		hooks = append(hooks, Hook{Type: PostToolUse, Tool: h.Tool, Command: h.Command})
	}
	return &Runner{hooks: hooks}
}

// RunPre executes all pre_tool_use hooks matching the tool name.
func (r *Runner) RunPre(ctx context.Context, toolName string, input json.RawMessage) {
	r.run(ctx, PreToolUse, toolName, input)
}

// RunPost executes all post_tool_use hooks matching the tool name.
func (r *Runner) RunPost(ctx context.Context, toolName string, input json.RawMessage) {
	r.run(ctx, PostToolUse, toolName, input)
}

func (r *Runner) run(ctx context.Context, hookType HookType, toolName string, input json.RawMessage) {
	for _, h := range r.hooks {
		if h.Type != hookType {
			continue
		}
		if h.Tool != "" && h.Tool != toolName {
			continue
		}

		hookCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		cmd := exec.CommandContext(hookCtx, "sh", "-c", h.Command)
		cmd.Env = append(os.Environ(),
			"TOOL_NAME="+toolName,
			"TOOL_INPUT="+string(input),
		)
		var out bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &out
		err := cmd.Run()
		cancel()
		if err != nil && hookCtx.Err() == context.DeadlineExceeded {
			fmt.Fprintf(os.Stderr, "[hooks] warning: hook %q timed out for tool %q\n", h.Command, toolName)
		}
		// hooks are best-effort; non-timeout failures are silent
	}
}
