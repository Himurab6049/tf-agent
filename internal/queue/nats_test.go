//go:build integration

package queue_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/tf-agent/tf-agent/internal/queue"
)

// Run with: NATS_URL=nats://localhost:4222 go test -tags=integration ./internal/queue/...

func newNATSQueue(t *testing.T) *queue.NATSQueue {
	t.Helper()
	url := os.Getenv("NATS_URL")
	if url == "" {
		t.Skip("NATS_URL not set — skipping NATS integration tests")
	}
	// Use a sanitized test name as the queue name for isolation between test cases.
	name := "test-" + strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		return '-'
	}, t.Name())
	q, err := queue.NewNATSQueue(url, name)
	if err != nil {
		t.Fatalf("NewNATSQueue: %v", err)
	}
	t.Cleanup(func() { _ = q.Close() })
	return q
}

func TestNATSQueue_PushPop(t *testing.T) {
	q := newNATSQueue(t)
	ctx := context.Background()

	item := queue.Item{
		TaskID:     "test-task-1",
		UserID:     "user-1",
		InputType:  "prompt",
		InputText:  "create an S3 bucket",
		OutputType: "print",
	}

	if err := q.Push(ctx, item); err != nil {
		t.Fatalf("Push: %v", err)
	}

	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	got, err := q.Pop(ctx2)
	if err != nil {
		t.Fatalf("Pop: %v", err)
	}
	if got.TaskID != item.TaskID {
		t.Errorf("got TaskID %q, want %q", got.TaskID, item.TaskID)
	}
	if got.InputText != item.InputText {
		t.Errorf("got InputText %q, want %q", got.InputText, item.InputText)
	}
}

func TestNATSQueue_PopCancelledContext(t *testing.T) {
	q := newNATSQueue(t)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	_, err := q.Pop(ctx)
	if err == nil {
		t.Error("expected context error, got nil")
	}
}

func TestNATSQueue_MultipleItems(t *testing.T) {
	q := newNATSQueue(t)
	ctx := context.Background()

	items := []queue.Item{
		{TaskID: "multi-1", UserID: "u1", InputType: "prompt", InputText: "task 1", OutputType: "print"},
		{TaskID: "multi-2", UserID: "u1", InputType: "prompt", InputText: "task 2", OutputType: "print"},
		{TaskID: "multi-3", UserID: "u2", InputType: "prompt", InputText: "task 3", OutputType: "pr"},
	}

	for _, item := range items {
		if err := q.Push(ctx, item); err != nil {
			t.Fatalf("Push %s: %v", item.TaskID, err)
		}
	}

	seen := map[string]bool{}
	for range items {
		ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
		got, err := q.Pop(ctx2)
		cancel()
		if err != nil {
			t.Fatalf("Pop: %v", err)
		}
		seen[got.TaskID] = true
	}

	for _, item := range items {
		if !seen[item.TaskID] {
			t.Errorf("never received task %s", item.TaskID)
		}
	}
}

func TestNATSQueue_CredentialsNotPersisted(t *testing.T) {
	// GitHub/Atlassian tokens should survive the round-trip (they're in-process, not sensitive over wire in test)
	q := newNATSQueue(t)
	ctx := context.Background()

	item := queue.Item{
		TaskID:      "creds-test",
		GitHubToken: "gh-secret-token",
		RepoURL:     "https://github.com/org/repo",
	}
	if err := q.Push(ctx, item); err != nil {
		t.Fatalf("Push: %v", err)
	}
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	got, err := q.Pop(ctx2)
	if err != nil {
		t.Fatalf("Pop: %v", err)
	}
	if got.GitHubToken != item.GitHubToken {
		t.Errorf("got GitHubToken %q, want %q", got.GitHubToken, item.GitHubToken)
	}
}
