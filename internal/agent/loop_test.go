package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/tf-agent/tf-agent/internal/commands"
	"github.com/tf-agent/tf-agent/internal/config"
	"github.com/tf-agent/tf-agent/internal/hooks"
	"github.com/tf-agent/tf-agent/internal/llm"
	"github.com/tf-agent/tf-agent/internal/permissions"
	"github.com/tf-agent/tf-agent/internal/session"
	"github.com/tf-agent/tf-agent/internal/skills"
	"github.com/tf-agent/tf-agent/internal/tools"
)

// buildTestAgent wires up a minimal Agent for testing.
func buildTestAgent(t *testing.T, provider llm.Provider, toolReg *tools.Registry) *Agent {
	t.Helper()
	dir := t.TempDir()

	sess, err := session.New(dir, "")
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}

	cfg := config.Defaults()
	// Set all permissions to auto so tool calls aren't blocked.
	cfg.Permissions.Bash = "auto"
	cfg.Permissions.Write = "auto"
	cfg.Permissions.Edit = "auto"
	cfg.Permissions.Read = "auto"
	cfg.Permissions.Glob = "auto"
	cfg.Permissions.Grep = "auto"
	cfg.Permissions.Default = "auto"
	cfg.Agent.MaxTurns = 5
	cfg.Agent.MaxTokens = 1024

	perm := permissions.NewChecker(&cfg.Permissions)
	hookRunner := hooks.NewRunner(&cfg.Hooks)
	skillReg := skills.NewRegistry()
	cmdReg := commands.NewRegistry()

	if toolReg == nil {
		toolReg = tools.NewRegistry()
	}

	return NewAgent(provider, toolReg, skillReg, sess, perm, hookRunner, cmdReg, cfg, dir, "claude-sonnet-4-6", "")
}

func collectEvents(ch <-chan TurnEvent) []TurnEvent {
	var evs []TurnEvent
	for e := range ch {
		evs = append(evs, e)
	}
	return evs
}

func TestAgent_SimpleText(t *testing.T) {
	events := []llm.Event{
		{Type: llm.EventText, Delta: "Hello "},
		{Type: llm.EventText, Delta: "world!"},
		{Type: llm.EventStop, StopReason: "end_turn"},
	}
	provider := llm.NewMockProvider("mock", events)
	agent := buildTestAgent(t, provider, nil)

	evs := collectEvents(agent.RunTurn(context.Background(), "say hello"))

	var text strings.Builder
	for _, e := range evs {
		if e.Type == TurnEventText {
			text.WriteString(e.Text)
		}
	}
	if !strings.Contains(text.String(), "Hello world!") {
		t.Errorf("expected combined text 'Hello world!', got: %q", text.String())
	}

	// Verify the turn ends with a Done event.
	last := evs[len(evs)-1]
	if last.Type != TurnEventDone {
		t.Errorf("last event should be TurnEventDone, got type %d", last.Type)
	}
}

func TestAgent_ToolCall(t *testing.T) {
	toolInput := json.RawMessage(`{"command":"echo tool_executed"}`)

	// The mock emits one tool_use event then stop.
	// The agent will execute the tool, then call Stream again with the result.
	// The second call should return end_turn.
	callCount := 0
	firstCall := []llm.Event{
		{Type: llm.EventToolUse, ToolUse: &llm.ToolUseEvent{
			ID:    "call_1",
			Name:  "bash",
			Input: toolInput,
		}},
		{Type: llm.EventStop, StopReason: "tool_use"},
	}
	secondCall := []llm.Event{
		{Type: llm.EventText, Delta: "Done."},
		{Type: llm.EventStop, StopReason: "end_turn"},
	}

	// Build a provider that returns different events on each call.
	multiProvider := &multiCallMock{calls: [][]llm.Event{firstCall, secondCall}, current: &callCount}

	// Register the bash tool.
	reg := tools.NewRegistry()
	reg.Register(&tools.BashTool{})

	agent := buildTestAgent(t, multiProvider, reg)
	evs := collectEvents(agent.RunTurn(context.Background(), "run echo"))

	// We should have seen at least one ToolStart and one ToolEnd event.
	var sawToolStart, sawToolEnd bool
	for _, e := range evs {
		if e.Type == TurnEventToolStart {
			sawToolStart = true
		}
		if e.Type == TurnEventToolEnd {
			sawToolEnd = true
		}
	}
	if !sawToolStart {
		t.Error("expected TurnEventToolStart, not seen")
	}
	if !sawToolEnd {
		t.Error("expected TurnEventToolEnd, not seen")
	}
}

func TestAgent_PermissionDeny(t *testing.T) {
	toolInput := json.RawMessage(`{"command":"echo should_not_run"}`)

	events := []llm.Event{
		{Type: llm.EventToolUse, ToolUse: &llm.ToolUseEvent{
			ID:    "call_deny",
			Name:  "bash",
			Input: toolInput,
		}},
		{Type: llm.EventStop, StopReason: "tool_use"},
	}

	// Second call (after tool result) returns end_turn.
	secondCall := []llm.Event{
		{Type: llm.EventText, Delta: "Permission was denied."},
		{Type: llm.EventStop, StopReason: "end_turn"},
	}

	callCount := 0
	provider := &multiCallMock{calls: [][]llm.Event{events, secondCall}, current: &callCount}

	reg := tools.NewRegistry()
	reg.Register(&tools.BashTool{})

	dir := t.TempDir()
	sess, err := session.New(dir, "")
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}

	cfg := config.Defaults()
	// Deny bash specifically.
	cfg.Permissions.Bash = "deny"
	cfg.Permissions.Default = "auto"
	cfg.Agent.MaxTurns = 5
	cfg.Agent.MaxTokens = 1024

	perm := permissions.NewChecker(&cfg.Permissions)
	hookRunner := hooks.NewRunner(&cfg.Hooks)
	skillReg := skills.NewRegistry()
	cmdReg := commands.NewRegistry()

	agent := NewAgent(provider, reg, skillReg, sess, perm, hookRunner, cmdReg, cfg, dir, "claude-sonnet-4-6", "")
	evs := collectEvents(agent.RunTurn(context.Background(), "run echo"))

	// The tool end event should mention permission denied, not the echo output.
	var toolEndSeen bool
	for _, e := range evs {
		if e.Type == TurnEventToolEnd && e.ToolResult != nil {
			toolEndSeen = true
			if !strings.Contains(e.ToolResult.Output, "Permission denied") {
				t.Errorf("expected 'Permission denied' in tool result, got: %q", e.ToolResult.Output)
			}
		}
	}
	if !toolEndSeen {
		t.Error("expected a TurnEventToolEnd event")
	}
}

// multiCallMock returns a different set of events for each successive Stream call.
type multiCallMock struct {
	calls   [][]llm.Event
	current *int
	name    string
}

func (m *multiCallMock) Stream(_ context.Context, req llm.Request) (<-chan llm.Event, error) {
	idx := *m.current
	*m.current++
	var evs []llm.Event
	if idx < len(m.calls) {
		evs = m.calls[idx]
	} else {
		evs = []llm.Event{{Type: llm.EventStop, StopReason: "end_turn"}}
	}
	ch := make(chan llm.Event, len(evs))
	for _, e := range evs {
		ch <- e
	}
	close(ch)
	return ch, nil
}

func (m *multiCallMock) Name() string { return "multi-mock" }

func TestAgent_ParallelToolExecution(t *testing.T) {
	cwd := t.TempDir()

	// Create a small file in the temp dir so glob and read have something to work with.
	testFile := cwd + "/hello.txt"
	if err := os.WriteFile(testFile, []byte("hello world\n"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	globInput := json.RawMessage(`{"pattern":"*.txt"}`)
	readInput := json.RawMessage(`{"file_path":"hello.txt"}`)

	callCount := 0
	firstCall := []llm.Event{
		{Type: llm.EventToolUse, ToolUse: &llm.ToolUseEvent{
			ID:    "call_glob",
			Name:  "glob",
			Input: globInput,
		}},
		{Type: llm.EventToolUse, ToolUse: &llm.ToolUseEvent{
			ID:    "call_read",
			Name:  "read",
			Input: readInput,
		}},
		{Type: llm.EventStop, StopReason: "tool_use"},
	}
	secondCall := []llm.Event{
		{Type: llm.EventText, Delta: "done"},
		{Type: llm.EventStop, StopReason: "end_turn"},
	}

	provider := &multiCallMock{calls: [][]llm.Event{firstCall, secondCall}, current: &callCount}

	reg := tools.NewRegistry()
	reg.Register(tools.NewGlobTool(cwd))
	reg.Register(tools.NewReadTool(cwd))

	agent := buildTestAgent(t, provider, reg)
	evs := collectEvents(agent.RunTurn(context.Background(), "parallel task"))

	var toolStarts, toolEnds []string
	for _, e := range evs {
		if e.Type == TurnEventToolStart && e.ToolCall != nil {
			toolStarts = append(toolStarts, e.ToolCall.ID)
		}
		if e.Type == TurnEventToolEnd && e.ToolResult != nil {
			toolEnds = append(toolEnds, e.ToolResult.ID)
		}
	}

	if len(toolStarts) != 2 {
		t.Errorf("expected 2 TurnEventToolStart events, got %d: %v", len(toolStarts), toolStarts)
	}
	if len(toolEnds) != 2 {
		t.Errorf("expected 2 TurnEventToolEnd events, got %d: %v", len(toolEnds), toolEnds)
	}

	// Both tool IDs must appear in starts.
	wantIDs := map[string]bool{"call_glob": false, "call_read": false}
	for _, id := range toolStarts {
		wantIDs[id] = true
	}
	for id, seen := range wantIDs {
		if !seen {
			t.Errorf("tool ID %q not seen in ToolStart events", id)
		}
	}

	// Both tool IDs must appear in ends.
	wantIDs = map[string]bool{"call_glob": false, "call_read": false}
	for _, id := range toolEnds {
		wantIDs[id] = true
	}
	for id, seen := range wantIDs {
		if !seen {
			t.Errorf("tool ID %q not seen in ToolEnd events", id)
		}
	}

	// Last event must be TurnEventDone.
	last := evs[len(evs)-1]
	if last.Type != TurnEventDone {
		t.Errorf("last event should be TurnEventDone, got type %d", last.Type)
	}
}

// ---------------------------------------------------------------------------
// isRetryable tests
// ---------------------------------------------------------------------------

func TestIsRetryable_Nil(t *testing.T) {
	if isRetryable(nil) {
		t.Error("nil error should not be retryable")
	}
}

func TestIsRetryable_Timeout(t *testing.T) {
	if !isRetryable(fmt.Errorf("request timeout after 30s")) {
		t.Error("timeout error should be retryable")
	}
}

func TestIsRetryable_429(t *testing.T) {
	if !isRetryable(fmt.Errorf("HTTP 429 too many requests")) {
		t.Error("429 error should be retryable")
	}
}

func TestIsRetryable_529(t *testing.T) {
	if !isRetryable(fmt.Errorf("upstream returned 529")) {
		t.Error("529 error should be retryable")
	}
}

func TestIsRetryable_503(t *testing.T) {
	if !isRetryable(fmt.Errorf("service unavailable: 503")) {
		t.Error("503 error should be retryable")
	}
}

func TestIsRetryable_Overloaded(t *testing.T) {
	if !isRetryable(fmt.Errorf("server is overloaded, try again later")) {
		t.Error("overloaded error should be retryable")
	}
}

func TestIsRetryable_RateLimit(t *testing.T) {
	if !isRetryable(fmt.Errorf("rate limit exceeded")) {
		t.Error("rate limit error should be retryable")
	}
}

func TestIsRetryable_TooManyRequests(t *testing.T) {
	if !isRetryable(fmt.Errorf("Too Many Requests")) {
		t.Error("'too many requests' error should be retryable")
	}
}

func TestIsRetryable_PermanentError(t *testing.T) {
	if isRetryable(fmt.Errorf("invalid API key")) {
		t.Error("permanent auth error should not be retryable")
	}
}

func TestIsRetryable_CaseSensitivity(t *testing.T) {
	// isRetryable uses strings.ToLower so mixed case should still match.
	if !isRetryable(fmt.Errorf("TIMEOUT connecting to endpoint")) {
		t.Error("uppercase TIMEOUT should be retryable")
	}
}

// ---------------------------------------------------------------------------

func TestAgent_ExecuteTools_ContextCancel(t *testing.T) {
	events := []llm.Event{
		{Type: llm.EventText, Delta: "starting"},
		{Type: llm.EventStop, StopReason: "end_turn"},
	}
	provider := llm.NewMockProvider("mock", events)
	agent := buildTestAgent(t, provider, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before RunTurn

	evs := collectEvents(agent.RunTurn(ctx, "anything"))

	// collectEvents must return (channel must close); verify TurnEventDone is present
	// or the slice is simply non-nil (no hang is the real requirement).
	// The agent may emit TurnEventDone or may return early — either is correct.
	_ = evs // no hang == pass
}
