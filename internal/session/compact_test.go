package session

import (
	"strings"
	"testing"
)

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("got %q, want %q", got, "hello...")
	}
	if got := truncate("", 10); got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestCompact_NothingToCompact(t *testing.T) {
	records := []Record{
		{Type: "user", Content: "hello"},
		{Type: "assistant", Content: "hi"},
	}
	out := Compact(records, 5)
	if len(out) != 2 {
		t.Errorf("expected 2 records unchanged, got %d", len(out))
	}
}

func TestCompact_EmptyRecords(t *testing.T) {
	out := Compact(nil, 5)
	if len(out) != 0 {
		t.Errorf("expected 0 records, got %d", len(out))
	}
}

func TestCompact_Triggers(t *testing.T) {
	// 4 user turns, keepTurns=2 → compact 2 oldest
	records := []Record{
		{Type: "user", Content: "turn1"},
		{Type: "assistant", Content: "resp1"},
		{Type: "user", Content: "turn2"},
		{Type: "assistant", Content: "resp2"},
		{Type: "user", Content: "turn3"},
		{Type: "assistant", Content: "resp3"},
		{Type: "user", Content: "turn4"},
		{Type: "assistant", Content: "resp4"},
	}

	out := Compact(records, 2)

	// First record should be the summary
	if out[0].Type != "user" {
		t.Errorf("first record should be user type summary, got %q", out[0].Type)
	}
	if !strings.HasPrefix(out[0].Content, "Previous conversation summary:") {
		t.Errorf("expected summary prefix, got: %q", out[0].Content)
	}

	// Remaining records should be the last keepTurns user turns + their responses
	found := false
	for _, r := range out {
		if r.Content == "turn3" {
			found = true
		}
	}
	if !found {
		t.Error("expected turn3 to be in kept records")
	}
}

func TestCompact_DefaultKeepTurns(t *testing.T) {
	// keepTurns=0 should default to 5
	var records []Record
	for i := 0; i < 4; i++ {
		records = append(records, Record{Type: "user", Content: "u"})
		records = append(records, Record{Type: "assistant", Content: "a"})
	}
	out := Compact(records, 0)
	// 4 user turns < default 5, so nothing compacted
	if len(out) != len(records) {
		t.Errorf("expected no compaction with 4 turns and default keepTurns=5, got %d records", len(out))
	}
}

func TestCompact_SummaryIncludesToolCalls(t *testing.T) {
	records := []Record{
		{Type: "user", Content: "do it"},
		{Type: "tool_use", Name: "bash", Content: "ls"},
		{Type: "tool_result", Name: "bash", Content: "file.tf"},
		{Type: "assistant", Content: "done"},
		{Type: "user", Content: "turn2"},
		{Type: "assistant", Content: "resp2"},
		{Type: "user", Content: "turn3"},
		{Type: "assistant", Content: "resp3"},
		{Type: "user", Content: "turn4"},
		{Type: "assistant", Content: "resp4"},
		{Type: "user", Content: "turn5"},
		{Type: "assistant", Content: "resp5"},
		{Type: "user", Content: "turn6"},
		{Type: "assistant", Content: "resp6"},
	}

	out := Compact(records, 3)
	summary := out[0].Content
	if !strings.Contains(summary, "[Tool bash called]") {
		t.Errorf("expected tool_use in summary, got: %q", summary)
	}
	if !strings.Contains(summary, "[Tool bash result:") {
		t.Errorf("expected tool_result in summary, got: %q", summary)
	}
}
