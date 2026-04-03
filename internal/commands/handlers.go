package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tf-agent/tf-agent/internal/config"
	"github.com/tf-agent/tf-agent/internal/session"
)

// SkillNamer is satisfied by skills.Registry.
type SkillNamer interface {
	Names() []string
}

// RegisterAll adds built-in slash command handlers to the registry.
func RegisterAll(
	r *Registry,
	sess *session.Store,
	model *string,
	totalIn, totalOut *int,
	clearFn func(),
	cfg *config.Config,
	skillReg SkillNamer,
) {
	r.Register("help", func(_ string) (string, error) {
		return `/help     — show this help
/clear    — clear conversation history
/compact  — summarize old turns to reduce context
/model    — show current model
/tokens   — show token usage
/cost     — show estimated session cost
/memory   — show loaded AGENT.md content
/sessions — list recent sessions
/export   — export session to JSON
/diff     — show git diff --stat
/provider — show current provider and model
/plan     — toggle plan mode (reasoning only, no tool execution)
/skills   — list available skills`, nil
	})

	r.Register("clear", func(_ string) (string, error) {
		clearFn()
		return "Conversation history cleared.", nil
	})

	r.Register("compact", func(_ string) (string, error) {
		records := sess.Records()
		compacted := session.Compact(records, 5)
		sess.Clear()
		for _, r := range compacted {
			_ = sess.Append(r)
		}
		return fmt.Sprintf("Context compacted: %d records -> %d records", len(records), len(compacted)), nil
	})

	r.Register("model", func(_ string) (string, error) {
		return "Current model: " + *model, nil
	})

	r.Register("tokens", func(_ string) (string, error) {
		return fmt.Sprintf("Tokens: %d in, %d out (total %d)", *totalIn, *totalOut, *totalIn+*totalOut), nil
	})

	r.Register("cost", func(_ string) (string, error) {
		// Approximate pricing for claude-sonnet-4-6:
		// $3/M input tokens, $15/M output tokens
		inCost := float64(*totalIn) / 1_000_000 * 3.0
		outCost := float64(*totalOut) / 1_000_000 * 15.0
		return fmt.Sprintf("Estimated cost: $%.4f (%d in @ $3/M, %d out @ $15/M)",
			inCost+outCost, *totalIn, *totalOut), nil
	})

	r.Register("memory", func(_ string) (string, error) {
		return "AGENT.md is loaded into the system prompt automatically. Edit ~/.tf-agent/AGENT.md or <cwd>/AGENT.md.", nil
	})

	r.Register("export", func(_ string) (string, error) {
		records := sess.Records()
		var sb strings.Builder
		for _, rec := range records {
			fmt.Fprintf(&sb, "[%s] %s: %s\n", rec.Timestamp.Format("15:04:05"), rec.Type, truncate(rec.Content, 200))
		}
		return sb.String(), nil
	})

	r.Register("diff", handleDiff)

	r.Register("provider", func(_ string) (string, error) {
		return handleProvider(cfg), nil
	})

	r.Register("sessions", func(_ string) (string, error) {
		return handleSessions()
	})

	r.Register("plan", func(_ string) (string, error) {
		return "Plan mode: the agent will reason and outline steps without executing tools. (Not yet enforced — use this as a prompt prefix.)", nil
	})

	r.Register("skills", func(_ string) (string, error) {
		return handleSkills(skillReg), nil
	})
}

func handleDiff(_ string) (string, error) {
	out, _ := exec.Command("git", "diff", "--stat").Output()
	if len(out) == 0 {
		return "No changes.", nil
	}
	return string(out), nil
}

func handleProvider(cfg *config.Config) string {
	provider := "anthropic"
	model := "claude-sonnet-4-6"
	if cfg != nil {
		if cfg.Provider.Name != "" {
			provider = cfg.Provider.Name
		}
		if cfg.Provider.Model != "" {
			model = cfg.Provider.Model
		}
	}
	return fmt.Sprintf("Provider: %-12s  Model: %s", provider, model)
}

func handleSessions() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	sessDir := filepath.Join(homeDir, ".tf-agent", "sessions")

	entries, err := os.ReadDir(sessDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "No sessions found.", nil
		}
		return "", fmt.Errorf("cannot read sessions directory: %w", err)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%-36s  %s\n", "ID", "DATE")
	fmt.Fprintf(&sb, "%s  %s\n", strings.Repeat("-", 36), strings.Repeat("-", 10))

	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".jsonl") {
			continue
		}
		id := strings.TrimSuffix(name, ".jsonl")
		info, err := e.Info()
		date := "unknown"
		if err == nil {
			date = info.ModTime().Format("2006-01-02")
		}
		fmt.Fprintf(&sb, "%-36s  %s\n", id, date)
		count++
	}

	if count == 0 {
		return "No sessions found.", nil
	}
	return sb.String(), nil
}

func handleSkills(reg SkillNamer) string {
	if reg == nil {
		return "No skill registry available."
	}
	names := reg.Names()
	sort.Strings(names)
	if len(names) == 0 {
		return "No skills registered."
	}
	var sb strings.Builder
	for _, n := range names {
		fmt.Fprintf(&sb, "  %-20s\n", n)
	}
	return strings.TrimRight(sb.String(), "\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
