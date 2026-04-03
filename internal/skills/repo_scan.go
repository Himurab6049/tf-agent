package skills

import (
	_ "embed"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed prompts/repo_scan.md
var repoScanPrompt string

// RepoScanSkill scans a repository and returns a summary of its structure.
type RepoScanSkill struct{}

func (s *RepoScanSkill) Name() string                         { return "repo_scan" }
func (s *RepoScanSkill) IsReadOnly() bool                     { return true }
func (s *RepoScanSkill) IsDestructive(_ json.RawMessage) bool { return false }
func (s *RepoScanSkill) Prompt() string                       { return repoScanPrompt }

func (s *RepoScanSkill) Description() string {
	return "Scan a repository directory and return a structural summary of files and directories. Useful for understanding a codebase."
}

func (s *RepoScanSkill) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {
				"type": "string",
				"description": "Root directory to scan (default: current working directory)"
			},
			"max_depth": {
				"type": "integer",
				"description": "Maximum directory depth to traverse (default 3)"
			}
		}
	}`)
}

func (s *RepoScanSkill) Execute(_ context.Context, input json.RawMessage) (string, error) {
	var args struct {
		Path     string `json:"path"`
		MaxDepth int    `json:"max_depth"`
	}
	_ = json.Unmarshal(input, &args)

	root := args.Path
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("repo_scan: getwd: %w", err)
		}
	}

	maxDepth := args.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 3
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Repository: %s\n\n", root)

	err := walk(&sb, root, root, 0, maxDepth)
	if err != nil {
		return sb.String(), err
	}
	return sb.String(), nil
}

func walk(sb *strings.Builder, root, path string, depth, maxDepth int) error {
	if depth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	indent := strings.Repeat("  ", depth)
	for _, e := range entries {
		name := e.Name()
		// Skip hidden dirs/files at depth > 0 (allow .git at top level for info).
		if strings.HasPrefix(name, ".") && depth > 0 {
			continue
		}
		if name == "node_modules" || name == "vendor" || name == ".git" {
			if e.IsDir() {
				fmt.Fprintf(sb, "%s%s/ (skipped)\n", indent, name)
			}
			continue
		}

		rel, _ := filepath.Rel(root, filepath.Join(path, name))
		if e.IsDir() {
			fmt.Fprintf(sb, "%s%s/\n", indent, name)
			_ = walk(sb, root, filepath.Join(path, name), depth+1, maxDepth)
		} else {
			info, _ := e.Info()
			size := ""
			if info != nil {
				size = fmt.Sprintf(" (%d bytes)", info.Size())
			}
			_ = rel
			fmt.Fprintf(sb, "%s%s%s\n", indent, name, size)
		}
	}
	return nil
}
