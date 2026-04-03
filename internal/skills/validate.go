package skills

import (
	_ "embed"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

//go:embed prompts/validate.md
var validatePrompt string

// ValidateSkill runs tflint and/or terraform validate on a directory.
type ValidateSkill struct{}

func (s *ValidateSkill) Name() string                         { return "validate_terraform" }
func (s *ValidateSkill) IsReadOnly() bool                     { return true }
func (s *ValidateSkill) IsDestructive(_ json.RawMessage) bool { return false }
func (s *ValidateSkill) Prompt() string                       { return validatePrompt }

func (s *ValidateSkill) Description() string {
	return "Run tflint and terraform validate on Terraform files. Returns lint/validation output."
}

func (s *ValidateSkill) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Directory containing .tf files to validate"
			},
			"run_tflint": {
				"type": "boolean",
				"description": "Run tflint (default true)"
			},
			"run_terraform_validate": {
				"type": "boolean",
				"description": "Run terraform validate (default true)"
			}
		},
		"required": ["path"]
	}`)
}

func (s *ValidateSkill) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Path                 string `json:"path"`
		RunTflint            *bool  `json:"run_tflint"`
		RunTerraformValidate *bool  `json:"run_terraform_validate"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("validate_terraform: invalid input: %w", err)
	}

	runTflint := true
	if args.RunTflint != nil {
		runTflint = *args.RunTflint
	}
	runTFValidate := true
	if args.RunTerraformValidate != nil {
		runTFValidate = *args.RunTerraformValidate
	}

	ctx2, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	type result struct {
		label string
		out   string
	}

	var wg sync.WaitGroup
	ch := make(chan result, 2)

	if runTflint {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out, err := runCmd(ctx2, args.Path, "tflint", "--format=json")
			if err != nil {
				ch <- result{"tflint", "tflint not found or error: " + err.Error() + "\n"}
			} else {
				ch <- result{"tflint", out + "\n"}
			}
		}()
	}

	if runTFValidate {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// terraform init first (quiet).
			_, _ = runCmd(ctx2, args.Path, "terraform", "init", "-backend=false", "-no-color")
			out, err := runCmd(ctx2, args.Path, "terraform", "validate", "-json", "-no-color")
			if err != nil {
				ch <- result{"terraform validate", "terraform not found or error: " + err.Error() + "\n"}
			} else {
				ch <- result{"terraform validate", out + "\n"}
			}
		}()
	}

	wg.Wait()
	close(ch)

	// Collect results; preserve stable ordering (tflint before terraform validate).
	results := map[string]string{}
	for r := range ch {
		results[r.label] = r.out
	}

	var combined string
	if runTflint {
		combined += "=== tflint ===\n" + results["tflint"]
	}
	if runTFValidate {
		combined += "=== terraform validate ===\n" + results["terraform validate"]
	}
	if combined == "" {
		combined = "No validation tools ran."
	}
	return combined, nil
}

func runCmd(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}
