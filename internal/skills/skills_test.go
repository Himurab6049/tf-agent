package skills

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Registry ---

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(&ClarifierSkill{})

	s, ok := r.Get("clarifier")
	if !ok {
		t.Fatal("expected to find clarifier skill")
	}
	if s.Name() != "clarifier" {
		t.Errorf("Name = %q, want clarifier", s.Name())
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown skill")
	}
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	r.Register(&ClarifierSkill{})
	r.Register(&RepoScanSkill{})

	names := r.Names()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
}

func TestRegistry_Schemas(t *testing.T) {
	r := NewRegistry()
	r.Register(&ClarifierSkill{})
	r.Register(&ValidateSkill{})

	schemas := r.Schemas()
	if len(schemas) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(schemas))
	}
	for _, s := range schemas {
		if s.Name == "" {
			t.Error("schema name should not be empty")
		}
		if s.Description == "" {
			t.Error("schema description should not be empty")
		}
	}
}

func TestRegistry_AllPrompts(t *testing.T) {
	r := NewRegistry()
	r.Register(&ClarifierSkill{})
	r.Register(&GenerateSkill{})

	prompts := r.AllPrompts()
	if len(prompts) == 0 {
		t.Error("expected non-empty prompts map")
	}
	for name, p := range prompts {
		if p == "" {
			t.Errorf("skill %q returned empty prompt", name)
		}
	}
}

func TestRegistry_Execute_Unknown(t *testing.T) {
	r := NewRegistry()
	_, err := r.Execute(context.Background(), "unknown_skill", json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for unknown skill")
	}
}

func TestRegistry_Execute_Known(t *testing.T) {
	r := NewRegistry()
	r.Register(&ClarifierSkill{})

	input, _ := json.Marshal(map[string]any{
		"request":   "create terraform",
		"questions": []string{"Which cloud provider?"},
	})
	out, err := r.Execute(context.Background(), "clarifier", input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output")
	}
}

// --- ClarifierSkill ---

func TestClarifier_Execute_WithQuestions(t *testing.T) {
	s := &ClarifierSkill{}
	input, _ := json.Marshal(map[string]any{
		"request":   "create terraform for infra",
		"questions": []string{"Which cloud provider?", "Which region?", "What environment?"},
	})
	out, err := s.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "Which cloud provider?") {
		t.Errorf("output missing first question: %q", out)
	}
	if !strings.Contains(out, "create terraform for infra") {
		t.Errorf("output missing request context: %q", out)
	}
}

func TestClarifier_Execute_MaxThreeQuestions(t *testing.T) {
	s := &ClarifierSkill{}
	input, _ := json.Marshal(map[string]any{
		"request":   "test",
		"questions": []string{"Q1", "Q2", "Q3", "Q4", "Q5"},
	})
	out, err := s.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(out, "Q4") {
		t.Error("output should not include more than 3 questions")
	}
}

func TestClarifier_Execute_NoQuestions(t *testing.T) {
	s := &ClarifierSkill{}
	input, _ := json.Marshal(map[string]any{
		"request":   "test",
		"questions": []string{},
	})
	out, err := s.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output even with no questions")
	}
}

func TestClarifier_Execute_InvalidInput(t *testing.T) {
	s := &ClarifierSkill{}
	_, err := s.Execute(context.Background(), json.RawMessage(`not-valid-json`))
	if err == nil {
		t.Error("expected error for invalid JSON input")
	}
}

func TestClarifier_Metadata(t *testing.T) {
	s := &ClarifierSkill{}
	if s.Name() != "clarifier" {
		t.Errorf("Name = %q", s.Name())
	}
	if !s.IsReadOnly() {
		t.Error("clarifier should be read-only")
	}
	if s.IsDestructive(nil) {
		t.Error("clarifier should not be destructive")
	}
	if s.Prompt() == "" {
		t.Error("Prompt should be non-empty")
	}
	if s.Schema() == nil {
		t.Error("Schema should not be nil")
	}
}

// --- GenerateSkill ---

func TestGenerate_Execute_WritesFiles(t *testing.T) {
	dir := t.TempDir()
	s := NewGenerateSkill(dir)

	input, _ := json.Marshal(map[string]any{
		"files": map[string]string{
			"main.tf":      `resource "aws_s3_bucket" "main" {}`,
			"variables.tf": `variable "env" { type = string }`,
		},
	})
	out, err := s.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "main.tf") {
		t.Errorf("output missing main.tf: %q", out)
	}

	// Verify files were actually written.
	content, err := os.ReadFile(filepath.Join(dir, "main.tf"))
	if err != nil {
		t.Fatalf("main.tf not written: %v", err)
	}
	if !strings.Contains(string(content), "aws_s3_bucket") {
		t.Errorf("main.tf content wrong: %q", string(content))
	}
}

func TestGenerate_Execute_CustomOutputDir(t *testing.T) {
	baseDir := t.TempDir()
	subDir := filepath.Join(baseDir, "terraform", "s3")

	s := NewGenerateSkill("")
	input, _ := json.Marshal(map[string]any{
		"files":      map[string]string{"outputs.tf": `output "bucket_arn" {}`},
		"output_dir": subDir,
	})
	_, err := s.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if _, err := os.Stat(filepath.Join(subDir, "outputs.tf")); err != nil {
		t.Errorf("outputs.tf not created in custom dir: %v", err)
	}
}

func TestGenerate_Execute_EmptyFiles(t *testing.T) {
	s := NewGenerateSkill(t.TempDir())
	input, _ := json.Marshal(map[string]any{"files": map[string]string{}})
	_, err := s.Execute(context.Background(), input)
	if err == nil {
		t.Error("expected error for empty files map")
	}
}

func TestGenerate_Execute_InvalidInput(t *testing.T) {
	s := NewGenerateSkill(t.TempDir())
	_, err := s.Execute(context.Background(), json.RawMessage(`bad json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGenerate_Metadata(t *testing.T) {
	s := NewGenerateSkill("")
	if s.Name() != "generate_terraform" {
		t.Errorf("Name = %q", s.Name())
	}
	if s.IsReadOnly() {
		t.Error("generate should not be read-only")
	}
	if s.Prompt() == "" {
		t.Error("Prompt should be non-empty")
	}
}

// --- RepoScanSkill ---

func TestRepoScan_Execute_ReturnsStructure(t *testing.T) {
	dir := t.TempDir()

	// Create some files to scan.
	_ = os.WriteFile(filepath.Join(dir, "main.tf"), []byte(`resource "aws_vpc" "main" {}`), 0644)
	_ = os.WriteFile(filepath.Join(dir, "variables.tf"), []byte(`variable "region" {}`), 0644)
	subDir := filepath.Join(dir, "modules")
	_ = os.MkdirAll(subDir, 0755)
	_ = os.WriteFile(filepath.Join(subDir, "eks.tf"), []byte(`module "eks" {}`), 0644)

	s := &RepoScanSkill{}
	input, _ := json.Marshal(map[string]any{"path": dir})
	out, err := s.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !strings.Contains(out, "main.tf") {
		t.Errorf("output missing main.tf: %q", out)
	}
	if !strings.Contains(out, "variables.tf") {
		t.Errorf("output missing variables.tf: %q", out)
	}
	if !strings.Contains(out, "modules") {
		t.Errorf("output missing modules dir: %q", out)
	}
}

func TestRepoScan_Execute_DefaultsToCurrentDir(t *testing.T) {
	s := &RepoScanSkill{}
	// Empty input — defaults to CWD.
	out, err := s.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output")
	}
}

func TestRepoScan_Execute_MaxDepth(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "a", "b", "c", "d")
	_ = os.MkdirAll(deep, 0755)
	_ = os.WriteFile(filepath.Join(deep, "deep.tf"), []byte(""), 0644)

	s := &RepoScanSkill{}
	input, _ := json.Marshal(map[string]any{"path": dir, "max_depth": 1})
	out, err := s.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if strings.Contains(out, "deep.tf") {
		t.Error("max_depth=1 should not include files 4 levels deep")
	}
}

func TestRepoScan_Metadata(t *testing.T) {
	s := &RepoScanSkill{}
	if s.Name() != "repo_scan" {
		t.Errorf("Name = %q", s.Name())
	}
	if !s.IsReadOnly() {
		t.Error("repo_scan should be read-only")
	}
	if s.Prompt() == "" {
		t.Error("Prompt should be non-empty")
	}
}

// --- SecurityScanSkill (parsing only — checkov not required) ---

func TestParseCheckovOutput_PassFail(t *testing.T) {
	data := []byte(`{
		"results": {
			"passed_checks": [{"check_id":"CKV_AWS_1","resource":"aws_s3_bucket.main"}],
			"failed_checks": [
				{
					"check_id":"CKV_AWS_2",
					"resource":"aws_s3_bucket.main",
					"check":{"name":"Ensure S3 bucket has versioning"},
					"repo_file_path":"main.tf",
					"file_line_range":[10,15]
				}
			]
		},
		"summary": {"passed":1,"failed":1}
	}`)
	out, err := parseCheckovOutput(data)
	if err != nil {
		t.Fatalf("parseCheckovOutput: %v", err)
	}
	if !strings.Contains(out, "1 passed") {
		t.Errorf("output missing pass count: %q", out)
	}
	if !strings.Contains(out, "1 failed") {
		t.Errorf("output missing fail count: %q", out)
	}
	if !strings.Contains(out, "CKV_AWS_2") {
		t.Errorf("output missing failed check ID: %q", out)
	}
	if !strings.Contains(out, "versioning") {
		t.Errorf("output missing check name: %q", out)
	}
}

func TestParseCheckovOutput_ArrayFormat(t *testing.T) {
	// Checkov sometimes returns a JSON array.
	data := []byte(`[{"results":{"passed_checks":[],"failed_checks":[]},"summary":{"passed":5,"failed":0}}]`)
	out, err := parseCheckovOutput(data)
	if err != nil {
		t.Fatalf("parseCheckovOutput: %v", err)
	}
	if !strings.Contains(out, "5 passed") {
		t.Errorf("expected 5 passed, got: %q", out)
	}
}

func TestParseCheckovOutput_Empty(t *testing.T) {
	out, err := parseCheckovOutput([]byte{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output for empty input")
	}
}

func TestSecurityScan_Metadata(t *testing.T) {
	s := &SecurityScanSkill{}
	if s.Name() != "SecurityScan" {
		t.Errorf("Name = %q", s.Name())
	}
	if !s.IsReadOnly() {
		t.Error("SecurityScan should be read-only")
	}
	if s.Prompt() == "" {
		t.Error("Prompt should be non-empty")
	}
}

// --- CreatePRSkill (metadata + parseRepoURL) ---

func TestParseRepoURL_Valid(t *testing.T) {
	cases := []struct {
		input         string
		wantOwner     string
		wantRepo      string
	}{
		{"github.com/org/repo", "org", "repo"},
		{"https://github.com/org/repo", "org", "repo"},
		{"http://github.com/org/repo", "org", "repo"},
	}
	for _, c := range cases {
		owner, repo, err := parseRepoURL(c.input)
		if err != nil {
			t.Errorf("parseRepoURL(%q): unexpected error: %v", c.input, err)
			continue
		}
		if owner != c.wantOwner {
			t.Errorf("parseRepoURL(%q): owner = %q, want %q", c.input, owner, c.wantOwner)
		}
		if repo != c.wantRepo {
			t.Errorf("parseRepoURL(%q): repo = %q, want %q", c.input, repo, c.wantRepo)
		}
	}
}

func TestParseRepoURL_Invalid(t *testing.T) {
	cases := []string{"", "github.com/onlyone", "not-a-url"}
	for _, c := range cases {
		_, _, err := parseRepoURL(c)
		if err == nil {
			t.Errorf("parseRepoURL(%q): expected error, got nil", c)
		}
	}
}

func TestCreatePR_Metadata(t *testing.T) {
	s := &CreatePRSkill{}
	if s.Name() != "CreatePR" {
		t.Errorf("Name = %q", s.Name())
	}
	if s.IsReadOnly() {
		t.Error("CreatePR should not be read-only")
	}
	if s.Prompt() == "" {
		t.Error("Prompt should be non-empty")
	}
}

func TestCreatePR_MissingToken(t *testing.T) {
	s := &CreatePRSkill{}
	// No GITHUB_TOKEN in env, no context creds.
	input, _ := json.Marshal(map[string]any{
		"repo_url": "github.com/org/repo",
		"branch":   "tf-agent/test",
		"title":    "test PR",
		"body":     "test",
		"files":    map[string]string{"main.tf": ""},
	})
	_, err := s.Execute(context.Background(), input)
	if err == nil {
		t.Error("expected error when GITHUB_TOKEN is not set")
	}
	if !strings.Contains(err.Error(), "github_token") {
		t.Errorf("error should mention github_token: %v", err)
	}
}

// --- ValidateSkill (metadata only — tflint/terraform not required) ---

func TestValidate_Metadata(t *testing.T) {
	s := &ValidateSkill{}
	if s.Name() != "validate_terraform" {
		t.Errorf("Name = %q", s.Name())
	}
	if !s.IsReadOnly() {
		t.Error("validate should be read-only")
	}
	if s.Prompt() == "" {
		t.Error("Prompt should be non-empty")
	}
}

func TestValidate_Execute_InvalidInput(t *testing.T) {
	s := &ValidateSkill{}
	_, err := s.Execute(context.Background(), json.RawMessage(`not json`))
	if err == nil {
		t.Error("expected error for invalid input")
	}
}

func TestValidate_Execute_MissingPath(t *testing.T) {
	s := &ValidateSkill{}
	// Missing required path — tools won't be found so we get a soft error result.
	input, _ := json.Marshal(map[string]any{"path": t.TempDir()})
	out, err := s.Execute(context.Background(), input)
	// Should not return a hard error — tools just won't be available.
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}
	if out == "" {
		t.Error("expected non-empty output")
	}
}
