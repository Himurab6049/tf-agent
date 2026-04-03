package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAgentMD_NoneExist(t *testing.T) {
	dir := t.TempDir()
	got := LoadAgentMD(dir)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestLoadAgentMD_OnlyProject(t *testing.T) {
	dir := t.TempDir()
	content := "# Project instructions"
	if err := os.WriteFile(filepath.Join(dir, "AGENT.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	got := LoadAgentMD(dir)
	if got != content {
		t.Errorf("got %q, want %q", got, content)
	}
}

func TestLoadAgentMD_BothFiles(t *testing.T) {
	dir := t.TempDir()

	// Write project AGENT.md
	projectContent := "# Project"
	if err := os.WriteFile(filepath.Join(dir, "AGENT.md"), []byte(projectContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Temporarily override home to a temp dir so we can write a global AGENT.md
	fakeHome := t.TempDir()
	globalDir := filepath.Join(fakeHome, ".tf-agent")
	if err := os.MkdirAll(globalDir, 0755); err != nil {
		t.Fatal(err)
	}
	globalContent := "# Global"
	if err := os.WriteFile(filepath.Join(globalDir, "AGENT.md"), []byte(globalContent), 0644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("HOME", fakeHome)

	got := LoadAgentMD(dir)
	if got == "" {
		t.Fatal("expected non-empty result")
	}
	// Should contain both, separated by ---
	if got != globalContent+"\n\n---\n\n"+projectContent {
		t.Errorf("unexpected combined content: %q", got)
	}
}
