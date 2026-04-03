//go:build integration

package llm

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// TestIntegrationAnthropic_SimpleText verifies end-to-end streaming against
// the real Anthropic API. Run with:
//
//	go test -tags integration -run TestIntegrationAnthropic_SimpleText ./internal/llm/
//
// Requires ANTHROPIC_API_KEY to be set in the environment.
func TestIntegrationAnthropic_SimpleText(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set — skipping integration test")
	}

	provider := NewAnthropicProvider(apiKey, false)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := Request{
		Model:     "claude-sonnet-4-6",
		MaxTokens: 64,
		Messages: []Message{
			{
				Role: "user",
				Content: []ContentBlock{
					{Type: "text", Text: "Say only the word PONG and nothing else."},
				},
			},
		},
	}

	ch, err := provider.Stream(ctx, req)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	var text strings.Builder
	for ev := range ch {
		switch ev.Type {
		case EventText:
			text.WriteString(ev.Delta)
		case EventError:
			t.Fatalf("stream error: %v", ev.Err)
		}
	}

	got := text.String()
	if !strings.Contains(got, "PONG") {
		t.Errorf("expected response to contain 'PONG', got: %q", got)
	}
}
