package taskctx

import (
	"context"
	"testing"
)

func TestWithCredentials_RoundTrip(t *testing.T) {
	creds := Credentials{
		OutputType:      "pr",
		RepoURL:         "github.com/org/repo",
		GitHubToken:     "ghp_test",
		AtlassianToken:  "atl_test",
		AtlassianDomain: "example.atlassian.net",
		AtlassianEmail:  "user@example.com",
	}

	ctx := WithCredentials(context.Background(), creds)
	got, ok := FromContext(ctx)

	if !ok {
		t.Fatal("FromContext returned ok=false, expected ok=true")
	}
	if got.GitHubToken != creds.GitHubToken {
		t.Errorf("GitHubToken = %q, want %q", got.GitHubToken, creds.GitHubToken)
	}
	if got.AtlassianDomain != creds.AtlassianDomain {
		t.Errorf("AtlassianDomain = %q, want %q", got.AtlassianDomain, creds.AtlassianDomain)
	}
	if got.OutputType != creds.OutputType {
		t.Errorf("OutputType = %q, want %q", got.OutputType, creds.OutputType)
	}
}

func TestFromContext_Missing(t *testing.T) {
	_, ok := FromContext(context.Background())
	if ok {
		t.Error("expected ok=false for context without credentials")
	}
}

func TestWithCredentials_DoesNotMutateParent(t *testing.T) {
	parent := context.Background()
	child := WithCredentials(parent, Credentials{GitHubToken: "child-token"})

	_, ok := FromContext(parent)
	if ok {
		t.Error("parent context should not have credentials after child is created")
	}

	got, ok := FromContext(child)
	if !ok || got.GitHubToken != "child-token" {
		t.Errorf("child credentials not set correctly")
	}
}
