package session

import (
	"testing"
)

func TestStore_AppendAndRead(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	records := []Record{
		{Type: "user", Content: "hello"},
		{Type: "assistant", Content: "hi there"},
		{Type: "user", Content: "what time is it?"},
	}
	for _, r := range records {
		if err := s.Append(r); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}

	got := s.Records()
	if len(got) != len(records) {
		t.Fatalf("expected %d records, got %d", len(records), len(got))
	}
	for i, r := range got {
		if r.Type != records[i].Type {
			t.Errorf("record[%d].Type = %q, want %q", i, r.Type, records[i].Type)
		}
		if r.Content != records[i].Content {
			t.Errorf("record[%d].Content = %q, want %q", i, r.Content, records[i].Content)
		}
	}
}

func TestStore_Persistence(t *testing.T) {
	dir := t.TempDir()

	// Write records in the first store instance.
	id := "persist-test-session"
	s1, err := New(dir, id)
	if err != nil {
		t.Fatalf("New (first): %v", err)
	}
	if err := s1.Append(Record{Type: "user", Content: "persist me"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	if err := s1.Append(Record{Type: "assistant", Content: "persisted!"}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	// Re-open the same session by ID.
	s2, err := New(dir, id)
	if err != nil {
		t.Fatalf("New (second): %v", err)
	}
	got := s2.Records()
	if len(got) != 2 {
		t.Fatalf("expected 2 records after re-open, got %d", len(got))
	}
	if got[0].Content != "persist me" {
		t.Errorf("record[0].Content = %q, want %q", got[0].Content, "persist me")
	}
	if got[1].Content != "persisted!" {
		t.Errorf("record[1].Content = %q, want %q", got[1].Content, "persisted!")
	}
}

func TestStore_Title(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Append a user message — it should become the implicit title.
	if err := s.Append(Record{Type: "user", Content: "build me a terraform module"}); err != nil {
		t.Fatalf("Append: %v", err)
	}

	records := s.Records()
	if len(records) == 0 {
		t.Fatal("expected at least one record")
	}
	first := records[0]
	if first.Type != "user" {
		t.Errorf("first record type = %q, want 'user'", first.Type)
	}
	if first.Content != "build me a terraform module" {
		t.Errorf("first record content = %q", first.Content)
	}
}

func TestStore_ToMessages(t *testing.T) {
	dir := t.TempDir()
	s, err := New(dir, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_ = s.Append(Record{Type: "user", Content: "hello"})
	_ = s.Append(Record{Type: "assistant", Content: "world", Model: "claude-sonnet-4-6"})

	records := s.Records()
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// Verify the types are preserved (the agent's buildMessages converts these).
	if records[0].Type != "user" {
		t.Errorf("records[0].Type = %q, want 'user'", records[0].Type)
	}
	if records[1].Type != "assistant" {
		t.Errorf("records[1].Type = %q, want 'assistant'", records[1].Type)
	}
	if records[1].Model != "claude-sonnet-4-6" {
		t.Errorf("records[1].Model = %q, want 'claude-sonnet-4-6'", records[1].Model)
	}
}
