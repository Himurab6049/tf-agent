package permissions

// Level represents the permission level for a tool.
type Level string

const (
	LevelAuto Level = "auto" // Always execute without prompting
	LevelAsk  Level = "ask"  // Ask user before executing
	LevelDeny Level = "deny" // Always deny execution
)
