package tools

import (
	"context"
	"encoding/json"
	"testing"
)

// stubTool is a minimal Executable for testing the registry.
type stubTool struct {
	name     string
	readOnly bool
}

func (s *stubTool) Name() string                              { return s.name }
func (s *stubTool) Description() string                      { return "stub" }
func (s *stubTool) Schema() json.RawMessage                  { return json.RawMessage(`{}`) }
func (s *stubTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	return "ok:" + s.name, nil
}
func (s *stubTool) IsReadOnly() bool                         { return s.readOnly }
func (s *stubTool) IsDestructive(_ json.RawMessage) bool     { return !s.readOnly }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tool := &stubTool{name: "mytool"}
	r.Register(tool)

	got, ok := r.Get("mytool")
	if !ok {
		t.Fatal("expected tool to be found")
	}
	if got.Name() != "mytool" {
		t.Errorf("got name %q, want %q", got.Name(), "mytool")
	}
}

func TestRegistry_GetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected missing tool to return ok=false")
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "dup"})
	defer func() {
		if rec := recover(); rec == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	r.Register(&stubTool{name: "dup"})
}

func TestRegistry_All(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "a"})
	r.Register(&stubTool{name: "b"})
	r.Register(&stubTool{name: "c"})

	all := r.All()
	if len(all) != 3 {
		t.Errorf("expected 3 tools, got %d", len(all))
	}
}

func TestRegistry_Execute(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "echo"})

	out, err := r.Execute(context.Background(), "echo", json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok:echo" {
		t.Errorf("got %q, want %q", out, "ok:echo")
	}
}

func TestRegistry_ExecuteUnknownTool(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute(context.Background(), "ghost", json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected error for unknown tool")
	}
}

func TestRegistry_Schemas(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubTool{name: "t1"})
	r.Register(&stubTool{name: "t2"})

	schemas := r.Schemas()
	if len(schemas) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(schemas))
	}
}
