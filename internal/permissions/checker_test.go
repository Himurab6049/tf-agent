package permissions

import (
	"testing"

	"github.com/tf-agent/tf-agent/internal/config"
)

func defaultCfg() *config.PermissionsConfig {
	return &config.PermissionsConfig{
		Bash:    "ask",
		Write:   "ask",
		Edit:    "ask",
		Read:    "auto",
		Glob:    "auto",
		Grep:    "auto",
		Default: "ask",
	}
}

func TestChecker_AllowByDefault(t *testing.T) {
	checker := NewChecker(defaultCfg())

	// Read-only tools should be auto (always allow) by default config.
	for _, tool := range []string{"read", "glob", "grep"} {
		level := checker.Check(tool, nil)
		if level != LevelAuto {
			t.Errorf("tool %q: expected LevelAuto, got %q", tool, level)
		}
	}
}

func TestChecker_GlobalAllow(t *testing.T) {
	// When all tools are set to "auto", nothing should require asking.
	cfg := &config.PermissionsConfig{
		Bash:    "auto",
		Write:   "auto",
		Edit:    "auto",
		Read:    "auto",
		Glob:    "auto",
		Grep:    "auto",
		Default: "auto",
	}
	checker := NewChecker(cfg)

	for _, tool := range []string{"bash", "write", "edit", "read", "glob", "grep", "unknown_tool"} {
		level := checker.Check(tool, nil)
		if level != LevelAuto {
			t.Errorf("tool %q: expected LevelAuto with global auto config, got %q", tool, level)
		}
	}
}

func TestChecker_AskFn(t *testing.T) {
	checker := NewChecker(defaultCfg())

	// Destructive tools (bash, write, edit) should require asking.
	for _, tool := range []string{"bash", "write", "edit"} {
		level := checker.Check(tool, nil)
		if level != LevelAsk {
			t.Errorf("tool %q: expected LevelAsk, got %q", tool, level)
		}
	}
}

func TestChecker_DenyLevel(t *testing.T) {
	cfg := &config.PermissionsConfig{
		Bash:    "deny",
		Default: "ask",
	}
	checker := NewChecker(cfg)

	level := checker.Check("bash", nil)
	if level != LevelDeny {
		t.Errorf("expected LevelDeny for bash, got %q", level)
	}

	// Other tools fall back to default.
	level = checker.Check("read", nil)
	if level != LevelAsk {
		t.Errorf("expected LevelAsk for read (default), got %q", level)
	}
}

func TestChecker_UnknownToolUsesDefault(t *testing.T) {
	cfg := &config.PermissionsConfig{
		Default: "deny",
	}
	checker := NewChecker(cfg)

	level := checker.Check("some_unknown_tool", nil)
	if level != LevelDeny {
		t.Errorf("unknown tool: expected default LevelDeny, got %q", level)
	}
}
