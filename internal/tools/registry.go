package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tf-agent/tf-agent/internal/llm"
)

// Executable is the interface all tools must implement.
type Executable interface {
	Name() string
	Description() string
	Schema() json.RawMessage        // JSON Schema for input
	Execute(ctx context.Context, input json.RawMessage) (string, error)
	IsReadOnly() bool
	IsDestructive(input json.RawMessage) bool
}

// Registry holds all registered tools.
type Registry struct {
	tools map[string]Executable
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Executable)}
}

// Register adds a tool. Panics on duplicate names.
func (r *Registry) Register(t Executable) {
	if _, exists := r.tools[t.Name()]; exists {
		panic(fmt.Sprintf("tool already registered: %s", t.Name()))
	}
	r.tools[t.Name()] = t
}

// Get returns a tool by name.
func (r *Registry) Get(name string) (Executable, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// All returns all registered tools.
func (r *Registry) All() []Executable {
	out := make([]Executable, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}

// Schemas returns the LLM tool schemas for all registered tools.
func (r *Registry) Schemas() []llm.ToolSchema {
	out := make([]llm.ToolSchema, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, llm.ToolSchema{
			Name:        t.Name(),
			Description: t.Description(),
			InputSchema: t.Schema(),
		})
	}
	return out
}

// Execute runs the named tool with the given JSON input.
func (r *Registry) Execute(ctx context.Context, name string, input json.RawMessage) (string, error) {
	t, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return t.Execute(ctx, input)
}
