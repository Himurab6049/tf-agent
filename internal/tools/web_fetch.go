package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// WebFetchTool fetches a URL and returns the content as plain text.
type WebFetchTool struct {
	client *http.Client
}

func NewWebFetchTool() *WebFetchTool {
	return &WebFetchTool{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (t *WebFetchTool) Name() string                         { return "web_fetch" }
func (t *WebFetchTool) IsReadOnly() bool                     { return true }
func (t *WebFetchTool) IsDestructive(_ json.RawMessage) bool { return false }

func (t *WebFetchTool) Description() string {
	return "Fetch a URL and return its content as plain text. Useful for reading documentation or web pages."
}

func (t *WebFetchTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"url": {
				"type": "string",
				"description": "URL to fetch"
			}
		},
		"required": ["url"]
	}`)
}

func (t *WebFetchTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("web_fetch: invalid input: %w", err)
	}
	if args.URL == "" {
		return "", fmt.Errorf("web_fetch: url is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, args.URL, nil)
	if err != nil {
		return "", fmt.Errorf("web_fetch: %w", err)
	}
	req.Header.Set("User-Agent", "tf-agent/1.0")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("web_fetch: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 500_000))
	if err != nil {
		return "", fmt.Errorf("web_fetch: read body: %w", err)
	}

	content := string(body)
	// Simple HTML tag stripping for readability.
	content = stripHTMLTags(content)
	content = strings.TrimSpace(content)

	if len(content) > 50_000 {
		content = content[:50_000] + "\n... (truncated)"
	}
	return fmt.Sprintf("URL: %s\nStatus: %d\n\n%s", args.URL, resp.StatusCode, content), nil
}

func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, ch := range s {
		switch {
		case ch == '<':
			inTag = true
		case ch == '>':
			inTag = false
			result.WriteRune(' ')
		case !inTag:
			result.WriteRune(ch)
		}
	}
	// Collapse whitespace runs.
	lines := strings.Split(result.String(), "\n")
	var cleaned []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			cleaned = append(cleaned, l)
		}
	}
	return strings.Join(cleaned, "\n")
}
