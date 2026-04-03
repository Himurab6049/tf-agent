package skills

import (
	_ "embed"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tf-agent/tf-agent/internal/taskctx"
)

//go:embed prompts/jira_fetch.md
var jiraFetchPrompt string

// JiraFetchSkill fetches a Jira ticket and returns its content as structured text.
type JiraFetchSkill struct{}

func (s *JiraFetchSkill) Name() string                         { return "jira_fetch" }
func (s *JiraFetchSkill) IsReadOnly() bool                     { return true }
func (s *JiraFetchSkill) IsDestructive(_ json.RawMessage) bool { return false }
func (s *JiraFetchSkill) Prompt() string                       { return jiraFetchPrompt }

func (s *JiraFetchSkill) Description() string {
	return "Fetch a Jira ticket by key and return its summary, description, and acceptance criteria."
}

func (s *JiraFetchSkill) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"ticket": {
				"type": "string",
				"description": "Jira ticket key, e.g. INFRA-123"
			},
			"domain": {
				"type": "string",
				"description": "Atlassian domain, e.g. mycompany.atlassian.net (optional if set in task context)"
			},
			"email": {
				"type": "string",
				"description": "Atlassian account email (optional if set in task context)"
			},
			"api_token": {
				"type": "string",
				"description": "Atlassian API token (optional if set in task context)"
			}
		},
		"required": ["ticket"]
	}`)
}

func (s *JiraFetchSkill) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Ticket   string `json:"ticket"`
		Domain   string `json:"domain"`
		Email    string `json:"email"`
		APIToken string `json:"api_token"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("jira_fetch: invalid input: %w", err)
	}
	if args.Ticket == "" {
		return "", fmt.Errorf("jira_fetch: ticket is required")
	}

	// Resolve credentials: context → input args → env vars.
	domain, email, token := args.Domain, args.Email, args.APIToken
	if creds, ok := taskctx.FromContext(ctx); ok {
		if creds.AtlassianDomain != "" {
			domain = creds.AtlassianDomain
		}
		if creds.AtlassianEmail != "" {
			email = creds.AtlassianEmail
		}
		if creds.AtlassianToken != "" {
			token = creds.AtlassianToken
		}
	}
	if domain == "" {
		domain = os.Getenv("ATLASSIAN_DOMAIN")
	}
	if email == "" {
		email = os.Getenv("ATLASSIAN_EMAIL")
	}
	if token == "" {
		token = os.Getenv("ATLASSIAN_TOKEN")
	}

	if domain == "" || email == "" || token == "" {
		return "", fmt.Errorf("jira_fetch: atlassian domain, email, and api_token are required")
	}

	url := fmt.Sprintf("https://%s/rest/api/3/issue/%s", domain, args.Ticket)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("jira_fetch: build request: %w", err)
	}
	req.SetBasicAuth(email, token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("jira_fetch: request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 500_000))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("jira_fetch: Jira API returned %d: %s", resp.StatusCode, string(body))
	}

	return parseJiraIssue(args.Ticket, body)
}

func parseJiraIssue(ticket string, data []byte) (string, error) {
	var issue struct {
		Key    string `json:"key"`
		Fields struct {
			Summary     string `json:"summary"`
			Status      struct{ Name string } `json:"status"`
			Priority    struct{ Name string } `json:"priority"`
			Labels      []string              `json:"labels"`
			Description *adfDoc               `json:"description"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(data, &issue); err != nil {
		return "", fmt.Errorf("jira_fetch: parse response: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# Jira Ticket: %s\n\n", issue.Key)
	fmt.Fprintf(&sb, "**Summary:** %s\n", issue.Fields.Summary)
	fmt.Fprintf(&sb, "**Status:** %s\n", issue.Fields.Status.Name)
	if issue.Fields.Priority.Name != "" {
		fmt.Fprintf(&sb, "**Priority:** %s\n", issue.Fields.Priority.Name)
	}
	if len(issue.Fields.Labels) > 0 {
		fmt.Fprintf(&sb, "**Labels:** %s\n", strings.Join(issue.Fields.Labels, ", "))
	}
	if issue.Fields.Description != nil {
		sb.WriteString("\n**Description:**\n")
		sb.WriteString(extractADFText(issue.Fields.Description))
	}

	return sb.String(), nil
}

// adfDoc is the Atlassian Document Format root node.
type adfDoc struct {
	Type    string          `json:"type"`
	Content []adfNode       `json:"content"`
	Text    string          `json:"text"`
	Attrs   map[string]any  `json:"attrs"`
}

type adfNode struct {
	Type    string    `json:"type"`
	Text    string    `json:"text"`
	Content []adfNode `json:"content"`
}

// extractADFText walks the ADF tree and extracts plain text.
func extractADFText(doc *adfDoc) string {
	if doc == nil {
		return ""
	}
	var sb strings.Builder
	for _, node := range doc.Content {
		writeADFNode(&sb, node, 0)
	}
	return sb.String()
}

func writeADFNode(sb *strings.Builder, node adfNode, depth int) {
	switch node.Type {
	case "text":
		sb.WriteString(node.Text)
	case "paragraph":
		for _, child := range node.Content {
			writeADFNode(sb, child, depth)
		}
		sb.WriteString("\n")
	case "heading":
		for _, child := range node.Content {
			writeADFNode(sb, child, depth)
		}
		sb.WriteString("\n")
	case "bulletList", "orderedList":
		for _, child := range node.Content {
			writeADFNode(sb, child, depth+1)
		}
	case "listItem":
		sb.WriteString(strings.Repeat("  ", depth) + "- ")
		for _, child := range node.Content {
			writeADFNode(sb, child, depth)
		}
	case "hardBreak":
		sb.WriteString("\n")
	case "codeBlock":
		sb.WriteString("```\n")
		for _, child := range node.Content {
			writeADFNode(sb, child, depth)
		}
		sb.WriteString("\n```\n")
	default:
		for _, child := range node.Content {
			writeADFNode(sb, child, depth)
		}
	}
}
