package llm

import (
	"context"
	"encoding/json"
)

// Provider is the interface all LLM backends must implement.
type Provider interface {
	Stream(ctx context.Context, req Request) (<-chan Event, error)
	Name() string
}

// Request is the unified request type sent to any provider.
type Request struct {
	Model     string
	System    string
	Messages  []Message
	Tools     []ToolSchema
	MaxTokens int
}

// Message is a single turn in the conversation.
type Message struct {
	Role    string         // "user" | "assistant"
	Content []ContentBlock
}

// ContentBlock represents one piece of content inside a message.
type ContentBlock struct {
	Type      string          // "text" | "tool_use" | "tool_result"
	Text      string
	ID        string
	Name      string
	Input     json.RawMessage
	ToolUseID string
	Content   string
}

// Event is emitted by the streaming channel.
type Event struct {
	Type       EventType
	Delta      string
	ToolUse    *ToolUseEvent
	Usage      *UsageEvent
	StopReason string
	Err        error
}

// EventType enumerates the kinds of events the stream can produce.
type EventType int

const (
	EventText    EventType = iota
	EventToolUse           // complete tool call ready to execute
	EventUsage             // token counts
	EventStop              // stream finished
	EventError             // unrecoverable error
)

// ToolUseEvent carries a completed tool call from the model.
type ToolUseEvent struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// UsageEvent carries token accounting data.
type UsageEvent struct {
	InputTokens  int
	OutputTokens int
	CacheRead    int
	CacheCreated int
}

// ToolSchema describes a tool that can be passed to the model.
type ToolSchema struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}
