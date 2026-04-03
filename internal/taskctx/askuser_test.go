package taskctx

import (
	"context"
	"testing"
)

func TestWithAskUser_Roundtrip(t *testing.T) {
	called := false
	fn := func(ctx context.Context, question string) (string, error) {
		called = true
		return "42", nil
	}

	ctx := WithAskUser(context.Background(), fn)
	answer, err := AskUser(ctx, "what is the answer?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if answer != "42" {
		t.Errorf("got %q, want %q", answer, "42")
	}
	if !called {
		t.Error("expected AskUserFunc to be called")
	}
}

func TestAskUser_NoopWhenNotSet(t *testing.T) {
	answer, err := AskUser(context.Background(), "hello?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if answer != "" {
		t.Errorf("expected empty answer, got %q", answer)
	}
}

func TestWithAskUser_DoesNotMutateParent(t *testing.T) {
	parent := context.Background()
	fn := func(ctx context.Context, q string) (string, error) { return "yes", nil }
	_ = WithAskUser(parent, fn)

	// Parent context should still have no AskUserFunc
	answer, err := AskUser(parent, "?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if answer != "" {
		t.Errorf("parent context was mutated, got answer %q", answer)
	}
}
