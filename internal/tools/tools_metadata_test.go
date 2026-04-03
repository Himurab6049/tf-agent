package tools

import (
	"testing"
)

func TestToolMetadata(t *testing.T) {
	cwd := t.TempDir()
	ts := []Executable{
		NewReadTool(cwd),
		NewWriteTool(cwd),
		NewEditTool(cwd),
		NewGlobTool(cwd),
		NewGrepTool(cwd),
		NewLsTool(cwd),
		&BashTool{},
		&AskUserTool{},
		&TaskTool{},
	}
	for _, tool := range ts {
		tool := tool
		t.Run(tool.Name(), func(t *testing.T) {
			if tool.Name() == "" {
				t.Error("Name() is empty")
			}
			if tool.Description() == "" {
				t.Error("Description() is empty")
			}
			schema := tool.Schema()
			if len(schema) == 0 {
				t.Error("Schema() is empty")
			}
			// Ensure these do not panic.
			_ = tool.IsReadOnly()
			_ = tool.IsDestructive(nil)
		})
	}
}
