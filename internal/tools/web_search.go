package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// WebSearchTool searches the web using Brave Search API or Serper API.
type WebSearchTool struct {
	client *http.Client
}

func NewWebSearchTool() *WebSearchTool {
	return &WebSearchTool{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (t *WebSearchTool) Name() string                         { return "WebSearch" }
func (t *WebSearchTool) IsReadOnly() bool                     { return true }
func (t *WebSearchTool) IsDestructive(_ json.RawMessage) bool { return false }

func (t *WebSearchTool) Description() string {
	return "Search the web using Brave Search or Serper. Requires BRAVE_API_KEY or SERPER_API_KEY."
}

func (t *WebSearchTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {
				"type": "string",
				"description": "Search query"
			},
			"count": {
				"type": "integer",
				"description": "Number of results to return (default 5, max 10)"
			}
		},
		"required": ["query"]
	}`)
}

func (t *WebSearchTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Query string `json:"query"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("WebSearch: invalid input: %w", err)
	}
	if args.Query == "" {
		return "", fmt.Errorf("WebSearch: query is required")
	}
	if args.Count <= 0 {
		args.Count = 5
	}
	if args.Count > 10 {
		args.Count = 10
	}

	braveKey := os.Getenv("BRAVE_API_KEY")
	serperKey := os.Getenv("SERPER_API_KEY")

	if braveKey == "" && serperKey == "" {
		return "WebSearch requires BRAVE_API_KEY or SERPER_API_KEY environment variable", nil
	}

	if braveKey != "" {
		return t.searchBrave(ctx, braveKey, args.Query, args.Count)
	}
	return t.searchSerper(ctx, serperKey, args.Query, args.Count)
}

type braveSearchResult struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

func (t *WebSearchTool) searchBrave(ctx context.Context, apiKey, query string, count int) (string, error) {
	endpoint := fmt.Sprintf("https://api.search.brave.com/res/v1/web/search?q=%s&count=%d",
		url.QueryEscape(query), count)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("WebSearch: build request: %w", err)
	}
	req.Header.Set("X-Subscription-Token", apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("WebSearch: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1_000_000))
	if err != nil {
		return "", fmt.Errorf("WebSearch: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("WebSearch: Brave API returned %d: %s", resp.StatusCode, string(body))
	}

	var result braveSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("WebSearch: parse response: %w", err)
	}

	return formatBraveResults(result), nil
}

func formatBraveResults(result braveSearchResult) string {
	if len(result.Web.Results) == 0 {
		return "No results found."
	}
	var buf bytes.Buffer
	for i, r := range result.Web.Results {
		fmt.Fprintf(&buf, "%d. %s\n   URL: %s\n   Snippet: %s\n\n", i+1, r.Title, r.URL, r.Description)
	}
	return buf.String()
}

type serperSearchResult struct {
	Organic []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic"`
}

func (t *WebSearchTool) searchSerper(ctx context.Context, apiKey, query string, count int) (string, error) {
	payload, err := json.Marshal(map[string]interface{}{
		"q":   query,
		"num": count,
	})
	if err != nil {
		return "", fmt.Errorf("WebSearch: build payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://google.serper.dev/search", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("WebSearch: build request: %w", err)
	}
	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("WebSearch: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1_000_000))
	if err != nil {
		return "", fmt.Errorf("WebSearch: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("WebSearch: Serper API returned %d: %s", resp.StatusCode, string(body))
	}

	var result serperSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("WebSearch: parse response: %w", err)
	}

	return formatSerperResults(result), nil
}

func formatSerperResults(result serperSearchResult) string {
	if len(result.Organic) == 0 {
		return "No results found."
	}
	var buf bytes.Buffer
	for i, r := range result.Organic {
		fmt.Fprintf(&buf, "%d. %s\n   URL: %s\n   Snippet: %s\n\n", i+1, r.Title, r.Link, r.Snippet)
	}
	return buf.String()
}
