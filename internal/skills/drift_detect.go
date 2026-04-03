package skills

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
	_ "embed"
)

//go:embed prompts/drift_detect.md
var driftDetectPrompt string

// DriftDetectSkill runs terraform plan to detect drift between live infra and IaC.
type DriftDetectSkill struct{}

func (s *DriftDetectSkill) Name() string                         { return "detect_drift" }
func (s *DriftDetectSkill) IsReadOnly() bool                     { return true }
func (s *DriftDetectSkill) IsDestructive(_ json.RawMessage) bool { return false }
func (s *DriftDetectSkill) Prompt() string                       { return driftDetectPrompt }

func (s *DriftDetectSkill) Description() string {
	return "Run terraform plan to detect drift between live infrastructure state and IaC. Returns a summary of drifted resources."
}

func (s *DriftDetectSkill) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Directory containing .tf files and initialized state"
			},
			"var_file": {
				"type": "string",
				"description": "Optional path to a .tfvars file"
			}
		},
		"required": ["path"]
	}`)
}

func (s *DriftDetectSkill) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Path    string `json:"path"`
		VarFile string `json:"var_file"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("detect_drift: invalid input: %w", err)
	}
	if args.Path == "" {
		return "", fmt.Errorf("detect_drift: path is required")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// terraform init (quiet, no backend migration)
	initOut, err := runDriftCmd(ctx, args.Path, "terraform", "init", "-input=false", "-no-color")
	if err != nil {
		return fmt.Sprintf("terraform init failed — cannot detect drift.\n%s\n%s", initOut, err.Error()), nil
	}

	// terraform plan -detailed-exitcode
	// Exit code 0 = no changes, 1 = error, 2 = changes (drift)
	planArgs := []string{"plan", "-detailed-exitcode", "-no-color", "-input=false"}
	if args.VarFile != "" {
		planArgs = append(planArgs, "-var-file="+args.VarFile)
	}

	planOut, planErr := runDriftCmd(ctx, args.Path, "terraform", planArgs...)

	if planErr == nil {
		return "No drift detected. Infrastructure matches IaC.\n\nterraform plan output:\n" + planOut, nil
	}

	// Check exit code: ExitError with code 2 means changes present (drift), not an error
	if exitErr, ok := planErr.(*exec.ExitError); ok {
		if exitErr.ExitCode() == 2 {
			return "Drift detected.\n\nterraform plan output:\n" + planOut, nil
		}
	}

	return fmt.Sprintf("terraform plan failed.\n%s\n%s", planOut, planErr.Error()), nil
}

func runDriftCmd(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}
