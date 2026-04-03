package hooks

// HookType distinguishes when a hook fires.
type HookType string

const (
	PreToolUse  HookType = "pre_tool_use"
	PostToolUse HookType = "post_tool_use"
)

// Hook is a configured shell command to run around tool execution.
type Hook struct {
	Type    HookType
	Tool    string // "" matches all tools
	Command string
}
