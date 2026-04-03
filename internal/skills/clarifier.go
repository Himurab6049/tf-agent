package skills

import (
	_ "embed"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tf-agent/tf-agent/internal/llm"
	"github.com/tf-agent/tf-agent/internal/taskctx"
)

//go:embed prompts/clarifier.md
var clarifierPrompt string

// ClarifierSkill makes an independent LLM call to decide what questions to ask,
// then pauses the task and asks the user each one via taskctx.AskUser.
// This mirrors the original Python POC: a self-contained clarifier that does not
// rely on the main agent to pre-generate questions.
type ClarifierSkill struct {
	provider llm.Provider
	model    string
}

func NewClarifierSkill(provider llm.Provider, model string) *ClarifierSkill {
	return &ClarifierSkill{provider: provider, model: model}
}

func (s *ClarifierSkill) Name() string                         { return "clarifier" }
func (s *ClarifierSkill) IsReadOnly() bool                     { return true }
func (s *ClarifierSkill) IsDestructive(_ json.RawMessage) bool { return false }
func (s *ClarifierSkill) Prompt() string                       { return clarifierPrompt }

func (s *ClarifierSkill) Description() string {
	return "Ask the user clarifying questions before writing any code. Pass the full task request and repo scan findings. The skill decides what to ask and pauses for user answers."
}

func (s *ClarifierSkill) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"request": {
				"type": "string",
				"description": "The full user request plus any relevant findings from the repo scan"
			}
		},
		"required": ["request"]
	}`)
}

func (s *ClarifierSkill) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Request   string    `json:"request"`
		Questions *[]string `json:"questions"` // nil = not provided (call LLM); non-nil = use as-is
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("clarifier: invalid input: %w", err)
	}

	var questions []string
	if args.Questions != nil {
		// Questions were explicitly provided — use them, skip LLM call.
		questions = *args.Questions
	} else {
		// No pre-provided questions — ask the LLM to generate them.
		var err error
		questions, err = s.getQuestions(ctx, args.Request)
		if err != nil {
			return "", fmt.Errorf("clarifier: LLM call failed: %w", err)
		}
	}

	if len(questions) == 0 {
		return "No clarifying questions needed — proceeding with available context.", nil
	}

	// Ask all questions in a single message.
	if len(questions) > 3 {
		questions = questions[:3]
	}
	var prompt strings.Builder
	prompt.WriteString("Before I proceed, I need a few details:\n\n")
	for i, q := range questions {
		fmt.Fprintf(&prompt, "%d. %s\n", i+1, q)
	}
	prompt.WriteString("\nPlease answer all questions in one reply.")

	answer, err := taskctx.AskUser(ctx, prompt.String())
	if err != nil {
		return "", fmt.Errorf("clarifier: ask_user failed: %w", err)
	}
	return fmt.Sprintf("Request: %s\n\nQuestions asked:\n%s\n\nUser response:\n%s", args.Request, prompt.String(), answer), nil
}

// getQuestions calls the LLM with the clarifier system prompt and returns
// a list of questions, or nil if the LLM responds with NO_QUESTIONS.
func (s *ClarifierSkill) getQuestions(ctx context.Context, request string) ([]string, error) {
	req := llm.Request{
		Model:  s.model,
		System: clarifierPrompt,
		Messages: []llm.Message{
			{
				Role: "user",
				Content: []llm.ContentBlock{
					{Type: "text", Text: request},
				},
			},
		},
		MaxTokens: 512,
	}

	eventCh, err := s.provider.Stream(ctx, req)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	for ev := range eventCh {
		if ev.Type == llm.EventError {
			return nil, ev.Err
		}
		if ev.Type == llm.EventText {
			sb.WriteString(ev.Delta)
		}
	}

	text := strings.TrimSpace(sb.String())
	if text == "" || strings.EqualFold(text, "NO_QUESTIONS") {
		return nil, nil
	}

	// Parse numbered list: "1. question\n2. question\n3. question"
	var questions []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip leading "1. " / "1) " / "- " prefixes.
		if len(line) > 2 && (line[1] == '.' || line[1] == ')') && line[0] >= '1' && line[0] <= '9' {
			line = strings.TrimSpace(line[2:])
		} else if strings.HasPrefix(line, "- ") {
			line = strings.TrimSpace(line[2:])
		}
		if line != "" {
			questions = append(questions, line)
		}
	}
	return questions, nil
}
