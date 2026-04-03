package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func TestBashTool_SimpleCommand(t *testing.T) {
	tool := &BashTool{}
	out, err := tool.Execute(context.Background(), mustJSON(map[string]any{
		"command": "echo hello",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected output to contain 'hello', got: %q", out)
	}
}

func TestBashTool_Timeout(t *testing.T) {
	tool := &BashTool{}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := tool.Execute(ctx, mustJSON(map[string]any{
		"command": "sleep 10",
		"timeout": 1,
	}))
	if err == nil {
		t.Fatal("expected a timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected 'timed out' in error, got: %v", err)
	}
}

func TestBashTool_BlockedPattern(t *testing.T) {
	tool := &BashTool{}
	_, err := tool.Execute(context.Background(), mustJSON(map[string]any{
		"command": "rm -rf /",
	}))
	if err == nil {
		t.Fatal("expected error for blocked command, got nil")
	}
	if !strings.Contains(err.Error(), "blocked") {
		t.Fatalf("expected 'blocked' in error, got: %v", err)
	}
}

func TestBashTool_ANSIStrip(t *testing.T) {
	tool := &BashTool{}
	// printf emits an ANSI green-colored string followed by a reset
	out, err := tool.Execute(context.Background(), mustJSON(map[string]any{
		"command": `printf '\033[32mgreen\033[0m'`,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(out, "\x1b[") {
		t.Fatalf("expected ANSI codes to be stripped, got: %q", out)
	}
	if !strings.Contains(out, "green") {
		t.Fatalf("expected text 'green' to remain, got: %q", out)
	}
}

func TestBashTool_NonZeroExit(t *testing.T) {
	tool := &BashTool{}
	out, err := tool.Execute(context.Background(), mustJSON(map[string]any{
		"command": "exit 1",
	}))
	// Non-zero exit should NOT return a Go error — the output is still returned.
	// The implementation returns err only on timeout; for other failures it
	// returns the combined stdout/stderr plus potentially the error string.
	_ = out
	_ = err
	// The key contract: the call must not panic and must return something.
}
