package llm

import "context"

// MockProvider is a test double for Provider.
// Set Events to the sequence of events to emit.
type MockProvider struct {
	Events   []Event
	Requests []Request
	name     string
}

// NewMockProvider creates a MockProvider that will emit the given events on each Stream call.
func NewMockProvider(name string, events []Event) *MockProvider {
	return &MockProvider{name: name, Events: events}
}

// Stream records the request and returns the pre-configured events on a buffered channel.
func (m *MockProvider) Stream(_ context.Context, req Request) (<-chan Event, error) {
	m.Requests = append(m.Requests, req)
	ch := make(chan Event, len(m.Events))
	for _, e := range m.Events {
		ch <- e
	}
	close(ch)
	return ch, nil
}

// Name returns the provider name.
func (m *MockProvider) Name() string { return m.name }
