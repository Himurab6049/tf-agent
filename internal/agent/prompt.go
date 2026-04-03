package agent

import (
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/tf-agent/tf-agent/internal/skills"
	"github.com/tf-agent/tf-agent/internal/tools"
)

// BuildSystemPrompt constructs the system prompt for the agent.
// It injects environment context, the autonomous pipeline definition,
// and per-skill guidance from each skill's embedded prompt.
func BuildSystemPrompt(cwd string, toolReg *tools.Registry, skillReg *skills.Registry, agentMD string) string {
	var sb strings.Builder

	// --- Identity ---
	sb.WriteString("You are tf-agent, an autonomous AI agent that generates production-quality Terraform infrastructure code.\n")
	sb.WriteString("You operate without human intervention unless critically blocked. You make reasonable decisions independently,\n")
	sb.WriteString("document your assumptions, and complete tasks end-to-end.\n")
	sb.WriteString("FORMATTING RULES (strictly enforced):\n")
	sb.WriteString("- Plain text only. No markdown of any kind.\n")
	sb.WriteString("- No bullet points (no - or * or + at line start)\n")
	sb.WriteString("- No bold (**text**), no italic (*text*), no headings (#), no backticks\n")
	sb.WriteString("- No emojis\n")
	sb.WriteString("- No quoted strings like \"example\" unless the user wrote them\n")
	sb.WriteString("- When listing examples, just write them one per line with no prefix\n\n")

	// --- Environment ---
	sb.WriteString("## Environment\n")
	fmt.Fprintf(&sb, "- OS: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Fprintf(&sb, "- Working directory: %s\n", cwd)
	fmt.Fprintf(&sb, "- Date: %s\n", time.Now().Format("2006-01-02"))
	if gitStatus := getGitStatus(cwd); gitStatus != "" {
		fmt.Fprintf(&sb, "- Git: %s\n", gitStatus)
	}
	sb.WriteString("\n")

	// --- Autonomous pipeline ---
	sb.WriteString("## Pipeline\n")
	sb.WriteString("For every infrastructure request, execute these steps in order without pausing for confirmation:\n\n")
	sb.WriteString("1. **repo_scan** — scan the repo for existing Terraform patterns, naming conventions, provider versions, and tag schemas\n")
	sb.WriteString("2. **clarifier** — ALWAYS call clarifier after repo_scan. Pass the user request plus key repo scan findings as the request argument. The skill will decide what to ask and pause for answers automatically.\n")
	sb.WriteString("3. **generate_terraform** — write HCL to disk split across providers.tf / variables.tf / main.tf / outputs.tf\n")
	sb.WriteString("4. **validate_terraform** — run tflint + terraform validate; fix errors and retry up to 2 times\n")
	sb.WriteString("5. **SecurityScan** — run checkov; fix blockers, document warnings in the PR body\n")
	sb.WriteString("6. **CreatePR** — open a GitHub PR with the generated files, validation results, and security findings\n\n")

	// --- Autonomy rules ---
	sb.WriteString("## Autonomy rules\n")
	sb.WriteString("- Proceed through all pipeline steps without asking for permission\n")
	sb.WriteString("- Make reasonable assumptions and document them in the PR body\n")
	sb.WriteString("- Use read/grep/glob/bash freely to gather context\n")
	sb.WriteString("- Stop and ask the user when: (a) clarifier step — always, for any field not in the request or repo scan, ")
	sb.WriteString("(b) validation failed after 2 retries, or (c) a security blocker cannot be auto-fixed\n")
	sb.WriteString("- Never ask \"should I proceed?\" — just proceed\n")
	sb.WriteString("- Prefer fixing issues automatically over reporting them and waiting\n\n")

	// --- Skill guidance ---
	if skillReg != nil {
		prompts := skillReg.AllPrompts()
		if len(prompts) > 0 {
			sb.WriteString("## Skill guidance\n")
			// Sort for deterministic output.
			names := make([]string, 0, len(prompts))
			for n := range prompts {
				names = append(names, n)
			}
			sort.Strings(names)
			for _, name := range names {
				sb.WriteString(prompts[name])
				sb.WriteString("\n\n")
			}
		}
	}

	// --- Available tools ---
	sb.WriteString("## Available tools\n")
	for _, t := range toolReg.All() {
		fmt.Fprintf(&sb, "- **%s**: %s\n", t.Name(), t.Description())
	}
	if skillReg != nil {
		for _, name := range skillReg.Names() {
			s, _ := skillReg.Get(name)
			fmt.Fprintf(&sb, "- **%s** (skill): %s\n", s.Name(), s.Description())
		}
	}
	sb.WriteString("\n")

	// --- Project instructions ---
	if agentMD != "" {
		sb.WriteString("## Project instructions (AGENT.md)\n")
		sb.WriteString(agentMD)
		sb.WriteString("\n")
	}

	return sb.String()
}

func getGitStatus(cwd string) string {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(out))
	if s == "" {
		return "clean"
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 5 {
		lines = lines[:5]
		lines = append(lines, fmt.Sprintf("...and %d more", len(strings.Split(s, "\n"))-5))
	}
	return strings.Join(lines, ", ")
}
