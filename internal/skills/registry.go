package skills

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tf-agent/tf-agent/internal/llm"
)

// Executable is the interface for skills (same contract as tools.Executable).
type Executable interface {
	Name() string
	Description() string
	// Prompt returns skill-specific instructions injected into the system prompt.
	// Return an empty string if no additional guidance is needed.
	Prompt() string
	Schema() json.RawMessage
	Execute(ctx context.Context, input json.RawMessage) (string, error)
	IsReadOnly() bool
	IsDestructive(input json.RawMessage) bool
}

// Registry holds all registered skills.
type Registry struct {
	skills map[string]Executable
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{skills: make(map[string]Executable)}
}

// Register adds a skill.
func (r *Registry) Register(s Executable) {
	r.skills[s.Name()] = s
}

// Get returns a skill by name.
func (r *Registry) Get(name string) (Executable, bool) {
	s, ok := r.skills[name]
	return s, ok
}

// Schemas returns LLM tool schemas for all skills.
func (r *Registry) Schemas() []llm.ToolSchema {
	out := make([]llm.ToolSchema, 0, len(r.skills))
	for _, s := range r.skills {
		out = append(out, llm.ToolSchema{
			Name:        s.Name(),
			Description: s.Description(),
			InputSchema: s.Schema(),
		})
	}
	return out
}

// Names returns the names of all registered skills.
func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.skills))
	for name := range r.skills {
		out = append(out, name)
	}
	return out
}

// AllPrompts returns a map of skill name to its Prompt() string (non-empty only).
func (r *Registry) AllPrompts() map[string]string {
	out := make(map[string]string, len(r.skills))
	for name, s := range r.skills {
		if p := s.Prompt(); p != "" {
			out[name] = p
		}
	}
	return out
}

// Execute runs the named skill.
func (r *Registry) Execute(ctx context.Context, name string, input json.RawMessage) (string, error) {
	s, ok := r.skills[name]
	if !ok {
		return "", fmt.Errorf("unknown skill: %s", name)
	}
	return s.Execute(ctx, input)
}
