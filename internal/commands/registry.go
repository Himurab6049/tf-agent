package commands

import (
	"strings"
)

// Handler is a function that handles a slash command.
// Returns the output string to display, or an error.
type Handler func(args string) (string, error)

// Registry holds slash command handlers.
type Registry struct {
	handlers map[string]Handler
}

// NewRegistry creates a new registry.
func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Handler)}
}

// Register adds a handler for a command name (without the leading slash).
func (r *Registry) Register(name string, h Handler) {
	r.handlers[name] = h
}

// Dispatch parses a slash command string and runs the matching handler.
// Returns (output, true) if handled, ("", false) if not a slash command.
func (r *Registry) Dispatch(input string) (string, bool) {
	if !strings.HasPrefix(input, "/") {
		return "", false
	}
	parts := strings.SplitN(strings.TrimPrefix(input, "/"), " ", 2)
	name := parts[0]
	args := ""
	if len(parts) > 1 {
		args = parts[1]
	}

	h, ok := r.handlers[name]
	if !ok {
		return "Unknown command: /" + name + " — try /help", true
	}
	out, err := h(args)
	if err != nil {
		return "Error: " + err.Error(), true
	}
	return out, true
}

// Names returns sorted list of registered command names.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.handlers))
	for n := range r.handlers {
		names = append(names, n)
	}
	return names
}
