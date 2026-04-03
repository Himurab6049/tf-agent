//go:build integration

package db_test

import (
	"context"
	"os"
	"testing"

	"github.com/tf-agent/tf-agent/internal/db"
)

// Run with: DB_URL=postgres://tfagent:tfagent@localhost:5432/tfagent?sslmode=disable go test -tags=integration ./internal/db/...

func newPostgresStore(t *testing.T) db.Store {
	t.Helper()
	url := os.Getenv("DB_URL")
	if url == "" {
		t.Skip("DB_URL not set — skipping postgres integration tests")
	}
	s, err := db.NewPostgres(url)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	// Truncate all test data before and after each test for isolation.
	s.TruncateForTest(t)
	t.Cleanup(func() {
		s.TruncateForTest(t)
		_ = s.Close()
	})
	return s
}

func TestPostgres_UserLifecycle(t *testing.T) {
	s := newPostgresStore(t)
	ctx := context.Background()

	u, err := s.CreateUser(ctx, "pg-test-user", "hash123", "member")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if u.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	got, err := s.GetUserByTokenHash(ctx, "hash123")
	if err != nil {
		t.Fatalf("GetUserByTokenHash: %v", err)
	}
	if got.Username != "pg-test-user" {
		t.Errorf("got username %q, want pg-test-user", got.Username)
	}

	if err := s.UpdateUsername(ctx, u.ID, "pg-test-user-renamed"); err != nil {
		t.Fatalf("UpdateUsername: %v", err)
	}

	if err := s.SetUserActive(ctx, u.ID, false); err != nil {
		t.Fatalf("SetUserActive: %v", err)
	}

	// Should not find inactive user by token.
	_, err = s.GetUserByTokenHash(ctx, "hash123")
	if err == nil {
		t.Error("expected error for inactive user, got nil")
	}

	if err := s.DeleteUser(ctx, u.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
}

func TestPostgres_EnsureAdmin(t *testing.T) {
	s := newPostgresStore(t)
	ctx := context.Background()

	if err := s.EnsureAdmin(ctx, "pg-admin", "adminhash"); err != nil {
		t.Fatalf("EnsureAdmin: %v", err)
	}
	// Idempotent — second call should update token hash.
	if err := s.EnsureAdmin(ctx, "pg-admin", "adminhash2"); err != nil {
		t.Fatalf("EnsureAdmin (2nd): %v", err)
	}
	u, err := s.GetUserByTokenHash(ctx, "adminhash2")
	if err != nil {
		t.Fatalf("GetUserByTokenHash: %v", err)
	}
	if u.Role != "admin" {
		t.Errorf("got role %q, want admin", u.Role)
	}
	_ = s.DeleteUser(ctx, u.ID)
}

func TestPostgres_TaskLifecycle(t *testing.T) {
	s := newPostgresStore(t)
	ctx := context.Background()

	u, err := s.CreateUser(ctx, "pg-task-user", "taskhash", "member")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { _ = s.DeleteUser(ctx, u.ID) })

	task, err := s.CreateTask(ctx, u.ID, "prompt", "create an S3 bucket", "pr")
	if err != nil {
		t.Fatalf("CreateTask: %v", err)
	}
	if task.Status != "queued" {
		t.Errorf("got status %q, want queued", task.Status)
	}

	if err := s.UpdateTaskStatus(ctx, task.ID, "running"); err != nil {
		t.Fatalf("UpdateTaskStatus running: %v", err)
	}

	got, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.Status != "running" {
		t.Errorf("got status %q, want running", got.Status)
	}
	if got.StartedAt == nil {
		t.Error("expected started_at to be set")
	}

	if err := s.UpdateTaskPendingQuestion(ctx, task.ID, "Which region?"); err != nil {
		t.Fatalf("UpdateTaskPendingQuestion: %v", err)
	}
	got, _ = s.GetTask(ctx, task.ID)
	if got.PendingQuestion != "Which region?" {
		t.Errorf("got pending_question %q, want 'Which region?'", got.PendingQuestion)
	}

	if err := s.UpdateTaskResult(ctx, task.ID, "done", "https://github.com/pr/1", "", "output text", 100, 200); err != nil {
		t.Fatalf("UpdateTaskResult: %v", err)
	}

	got, err = s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatalf("GetTask after done: %v", err)
	}
	if got.Status != "done" {
		t.Errorf("got status %q, want done", got.Status)
	}
	if got.InputTokens != 100 || got.OutputTokens != 200 {
		t.Errorf("token counts wrong: in=%d out=%d", got.InputTokens, got.OutputTokens)
	}
	if got.CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestPostgres_MarkStaleTasksFailed(t *testing.T) {
	s := newPostgresStore(t)
	ctx := context.Background()

	u, err := s.CreateUser(ctx, "pg-stale-user", "stalehash", "member")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { _ = s.DeleteUser(ctx, u.ID) })

	task, _ := s.CreateTask(ctx, u.ID, "prompt", "stale task", "print")
	_ = s.UpdateTaskStatus(ctx, task.ID, "running")

	if err := s.MarkStaleTasksFailed(ctx); err != nil {
		t.Fatalf("MarkStaleTasksFailed: %v", err)
	}

	got, _ := s.GetTask(ctx, task.ID)
	if got.Status != "failed" {
		t.Errorf("got status %q, want failed", got.Status)
	}
}

func TestPostgres_UserSettings(t *testing.T) {
	s := newPostgresStore(t)
	ctx := context.Background()

	u, err := s.CreateUser(ctx, "pg-settings-user", "settingshash", "member")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() { _ = s.DeleteUser(ctx, u.ID) })

	// Empty settings should return default.
	us, err := s.GetUserSettings(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetUserSettings (empty): %v", err)
	}
	if us.GitHubToken != "" {
		t.Errorf("expected empty github token, got %q", us.GitHubToken)
	}

	us.GitHubToken = "enc-gh-token"
	us.AtlassianDomain = "myco.atlassian.net"
	if err := s.UpsertUserSettings(ctx, us); err != nil {
		t.Fatalf("UpsertUserSettings: %v", err)
	}

	got, err := s.GetUserSettings(ctx, u.ID)
	if err != nil {
		t.Fatalf("GetUserSettings: %v", err)
	}
	if got.GitHubToken != "enc-gh-token" {
		t.Errorf("got github token %q, want enc-gh-token", got.GitHubToken)
	}
}

func TestPostgres_Ping(t *testing.T) {
	s := newPostgresStore(t)
	if err := s.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
