package session

import (
	"fmt"
	"strings"
)

// Compact summarises old turns to reduce context length.
// It keeps the last keepTurns pairs and summarises everything older.
func Compact(records []Record, keepTurns int) []Record {
	if keepTurns <= 0 {
		keepTurns = 5
	}

	// Count user turns.
	var userTurnIndexes []int
	for i, r := range records {
		if r.Type == "user" {
			userTurnIndexes = append(userTurnIndexes, i)
		}
	}

	if len(userTurnIndexes) <= keepTurns {
		return records // nothing to compact
	}

	cutIndex := userTurnIndexes[len(userTurnIndexes)-keepTurns]

	// Summarise everything before cutIndex.
	var sb strings.Builder
	sb.WriteString("Previous conversation summary:\n")
	for _, r := range records[:cutIndex] {
		switch r.Type {
		case "user":
			fmt.Fprintf(&sb, "User: %s\n", truncate(r.Content, 200))
		case "assistant":
			fmt.Fprintf(&sb, "Assistant: %s\n", truncate(r.Content, 200))
		case "tool_use":
			fmt.Fprintf(&sb, "[Tool %s called]\n", r.Name)
		case "tool_result":
			fmt.Fprintf(&sb, "[Tool %s result: %s]\n", r.Name, truncate(r.Content, 100))
		}
	}

	summary := Record{
		Type:    "user",
		Content: sb.String(),
	}

	result := []Record{summary}
	result = append(result, records[cutIndex:]...)
	return result
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
