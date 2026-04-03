package permissions

import (
	"encoding/json"
	"testing"

	"github.com/tf-agent/tf-agent/internal/config"
)

func TestChecker_SessionOverrideTakesPrecedence(t *testing.T) {
	cfg := &config.PermissionsConfig{
		Bash: "ask",
	}
	c := NewChecker(cfg)

	// Without override: bash should be "ask"
	if got := c.Check("bash", nil); got != LevelAsk {
		t.Errorf("expected LevelAsk, got %q", got)
	}

	// Set session override to deny
	c.session["bash"] = LevelDeny
	if got := c.Check("bash", nil); got != LevelDeny {
		t.Errorf("expected LevelDeny after session override, got %q", got)
	}
}

func TestChecker_SessionOverrideAuto(t *testing.T) {
	cfg := &config.PermissionsConfig{
		Bash: "ask",
	}
	c := NewChecker(cfg)
	c.session["bash"] = LevelAuto

	if got := c.Check("bash", nil); got != LevelAuto {
		t.Errorf("expected LevelAuto from session override, got %q", got)
	}
}

func TestChecker_UnknownToolFallsBackToDefault(t *testing.T) {
	cfg := &config.PermissionsConfig{
		Default: "auto",
	}
	c := NewChecker(cfg)

	if got := c.Check("unknown_tool", json.RawMessage(`{}`)); got != LevelAuto {
		t.Errorf("expected default LevelAuto for unknown tool, got %q", got)
	}
}

func TestChecker_PerToolConfigRespected(t *testing.T) {
	cfg := &config.PermissionsConfig{
		Bash:  "deny",
		Write: "ask",
		Read:  "auto",
	}
	c := NewChecker(cfg)

	cases := map[string]Level{
		"bash":  LevelDeny,
		"write": LevelAsk,
		"read":  LevelAuto,
	}
	for tool, want := range cases {
		if got := c.Check(tool, nil); got != want {
			t.Errorf("tool %q: expected %q, got %q", tool, want, got)
		}
	}
}
