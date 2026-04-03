package skills

import (
	_ "embed"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed prompts/generate.md
var generatePrompt string

// GenerateSkill writes generated Terraform HCL files to disk.
type GenerateSkill struct {
	outputDir string
}

func NewGenerateSkill(outputDir string) *GenerateSkill {
	return &GenerateSkill{outputDir: outputDir}
}

func (s *GenerateSkill) Name() string                         { return "generate_terraform" }
func (s *GenerateSkill) IsReadOnly() bool                     { return false }
func (s *GenerateSkill) IsDestructive(_ json.RawMessage) bool { return false }
func (s *GenerateSkill) Prompt() string                       { return generatePrompt }

func (s *GenerateSkill) Description() string {
	return "Write Terraform HCL files to disk. Provide a map of filename to HCL content."
}

func (s *GenerateSkill) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"files": {
				"type": "object",
				"description": "Map of filename to HCL content (e.g. {\"main.tf\": \"resource ...\"})",
				"additionalProperties": {"type": "string"}
			},
			"output_dir": {
				"type": "string",
				"description": "Directory to write files (default: current working directory)"
			}
		},
		"required": ["files"]
	}`)
}

func (s *GenerateSkill) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Files     map[string]string `json:"files"`
		OutputDir string            `json:"output_dir"`
	}
	if err := json.Unmarshal(input, &args); err != nil {
		return "", fmt.Errorf("generate_terraform: invalid input: %w", err)
	}
	if len(args.Files) == 0 {
		return "", fmt.Errorf("generate_terraform: files is required")
	}

	outDir := s.outputDir
	if args.OutputDir != "" {
		outDir = args.OutputDir
	}
	if outDir == "" {
		var err error
		outDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("generate_terraform: getwd: %w", err)
		}
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("generate_terraform: mkdir: %w", err)
	}

	var written []string
	for name, content := range args.Files {
		path := filepath.Join(outDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("generate_terraform: write %s: %w", name, err)
		}
		written = append(written, path)
	}

	return fmt.Sprintf("Generated %d Terraform file(s):\n%v", len(written), written), nil
}
