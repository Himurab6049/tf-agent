package permissions

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/tf-agent/tf-agent/internal/config"
)

// replaceStdin swaps os.Stdin with a pipe whose write end is returned. The
// caller must write the desired input and close the write end, then call the
// returned restore function to put the original stdin back.
func replaceStdin(t *testing.T) (write *os.File, restore func()) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	orig := os.Stdin
	os.Stdin = r
	return w, func() {
		os.Stdin = orig
		r.Close()
	}
}

// TestPrompt_AllowOnce verifies "y" returns LevelAuto without persisting a
// session override.
func TestPrompt_AllowOnce(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{Bash: "ask"})

	w, restore := replaceStdin(t)
	defer restore()

	io.WriteString(w, "y\n")
	w.Close()

	level := c.Prompt("bash", "echo hi")
	if level != LevelAuto {
		t.Errorf("expected LevelAuto for 'y', got %q", level)
	}
	// Must not be persisted as a session override.
	if _, ok := c.session["bash"]; ok {
		t.Error("session override should not be set after 'allow once'")
	}
}

// TestPrompt_DenyOnce verifies "n" returns LevelDeny without persisting a
// session override.
func TestPrompt_DenyOnce(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{Bash: "ask"})

	w, restore := replaceStdin(t)
	defer restore()

	io.WriteString(w, "n\n")
	w.Close()

	level := c.Prompt("bash", "rm -rf /tmp/x")
	if level != LevelDeny {
		t.Errorf("expected LevelDeny for 'n', got %q", level)
	}
	if _, ok := c.session["bash"]; ok {
		t.Error("session override should not be set after 'deny once'")
	}
}

// TestPrompt_AllowAlways verifies "a" returns LevelAuto and persists session
// override so that subsequent Check calls return LevelAuto without prompting.
func TestPrompt_AllowAlways(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{Bash: "ask"})

	w, restore := replaceStdin(t)
	defer restore()

	io.WriteString(w, "a\n")
	w.Close()

	level := c.Prompt("bash", "make build")
	if level != LevelAuto {
		t.Errorf("expected LevelAuto for 'a', got %q", level)
	}
	if got, ok := c.session["bash"]; !ok || got != LevelAuto {
		t.Errorf("session['bash'] = %q (ok=%v), want LevelAuto", got, ok)
	}

	// Subsequent Check must use the session override.
	if got := c.Check("bash", json.RawMessage(`{}`)); got != LevelAuto {
		t.Errorf("Check after 'always': expected LevelAuto, got %q", got)
	}
}

// TestPrompt_DenyAlways verifies "N" returns LevelDeny and persists session
// override so subsequent Check calls return LevelDeny.
func TestPrompt_DenyAlways(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{Bash: "ask"})

	w, restore := replaceStdin(t)
	defer restore()

	io.WriteString(w, "N\n")
	w.Close()

	level := c.Prompt("bash", "curl http://example.com")
	if level != LevelDeny {
		t.Errorf("expected LevelDeny for 'N', got %q", level)
	}
	if got, ok := c.session["bash"]; !ok || got != LevelDeny {
		t.Errorf("session['bash'] = %q (ok=%v), want LevelDeny", got, ok)
	}

	if got := c.Check("bash", nil); got != LevelDeny {
		t.Errorf("Check after 'never': expected LevelDeny, got %q", got)
	}
}

// TestPrompt_EmptyLineAllows verifies that pressing enter (empty input)
// behaves the same as "y" — allow once.
func TestPrompt_EmptyLineAllows(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{Bash: "ask"})

	w, restore := replaceStdin(t)
	defer restore()

	io.WriteString(w, "\n")
	w.Close()

	level := c.Prompt("bash", "ls")
	if level != LevelAuto {
		t.Errorf("expected LevelAuto for empty enter, got %q", level)
	}
}

// TestPrompt_InvalidThenValid verifies that an unrecognised answer re-prompts
// and the second (valid) answer is used.
func TestPrompt_InvalidThenValid(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{Write: "ask"})

	w, restore := replaceStdin(t)
	defer restore()

	// First line is garbage; second line is valid.
	io.WriteString(w, "what\ny\n")
	w.Close()

	level := c.Prompt("write", "write file.txt")
	if level != LevelAuto {
		t.Errorf("expected LevelAuto after invalid+valid input, got %q", level)
	}
}

// TestPrompt_StdinClosed verifies that closing stdin (EOF) causes Prompt to
// return LevelDeny (safe default).
func TestPrompt_StdinClosed(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{Bash: "ask"})

	w, restore := replaceStdin(t)
	defer restore()

	// Close without writing — immediate EOF.
	w.Close()

	level := c.Prompt("bash", "dangerous command")
	if level != LevelDeny {
		t.Errorf("expected LevelDeny on stdin EOF, got %q", level)
	}
}

// TestChecker_CheckNilInput verifies that nil json.RawMessage is handled
// gracefully by Check (it is explicitly discarded in the implementation).
func TestChecker_CheckNilInput(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{
		Read:    "auto",
		Default: "ask",
	})

	if got := c.Check("read", nil); got != LevelAuto {
		t.Errorf("nil input: expected LevelAuto for read, got %q", got)
	}
	if got := c.Check("bash", nil); got != LevelAsk {
		t.Errorf("nil input: expected LevelAsk for bash (default), got %q", got)
	}
}

// TestChecker_DefaultFallbackWhenEmpty verifies that an empty Default in
// config results in LevelAsk (the hard-coded fallback in NewChecker).
func TestChecker_DefaultFallbackWhenEmpty(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{}) // all fields empty
	if got := c.Check("unknown", nil); got != LevelAsk {
		t.Errorf("empty default: expected LevelAsk, got %q", got)
	}
}

// TestChecker_EmptyPerToolLevelUsesDefault verifies that a per-tool config
// value of "" is treated as unset and falls through to the default level.
func TestChecker_EmptyPerToolLevelUsesDefault(t *testing.T) {
	c := NewChecker(&config.PermissionsConfig{
		Bash:    "", // explicitly empty — should fall through
		Default: "auto",
	})
	if got := c.Check("bash", nil); got != LevelAuto {
		t.Errorf("empty per-tool: expected default LevelAuto, got %q", got)
	}
}
