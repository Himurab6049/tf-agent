package session

import (
	"os"
	"path/filepath"
)

// LoadAgentMD loads AGENT.md files hierarchically:
// ~/.tf-agent/AGENT.md (global), then cwd/AGENT.md (project).
// Returns combined content.
func LoadAgentMD(cwd string) string {
	var parts []string

	home, _ := os.UserHomeDir()
	globalPath := filepath.Join(home, ".tf-agent", "AGENT.md")
	if data, err := os.ReadFile(globalPath); err == nil {
		parts = append(parts, string(data))
	}

	projectPath := filepath.Join(cwd, "AGENT.md")
	if data, err := os.ReadFile(projectPath); err == nil {
		parts = append(parts, string(data))
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "\n\n---\n\n"
		}
		result += p
	}
	return result
}
