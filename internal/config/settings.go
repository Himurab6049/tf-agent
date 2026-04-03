package config

// Config is the top-level configuration structure loaded from
// ~/.tf-agent/config.toml and overridden by environment variables.
type Config struct {
	Provider    ProviderConfig    `toml:"provider"`
	Agent       AgentConfig       `toml:"agent"`
	Permissions PermissionsConfig `toml:"permissions"`
	Hooks       HooksConfig       `toml:"hooks"`
	Server      ServerConfig      `toml:"server"`
}

type ServerConfig struct {
	Port               int    `toml:"port"`
	PostgresURL        string `toml:"postgres_url"` // postgres://user:pass@host:5432/db?sslmode=disable
	QueueDriver        string `toml:"queue_driver"` // memory (default) | nats
	NatsURL            string `toml:"nats_url"`     // nats://host:4222
	LLMConcurrency     int    `toml:"llm_concurrency"`
	PerUserConcurrency int    `toml:"per_user_concurrency"`
	QueueBuffer        int    `toml:"queue_buffer"`
}

type ProviderConfig struct {
	Name      string            `toml:"name"`
	Model     string            `toml:"model"`
	Anthropic AnthropicConfig   `toml:"anthropic"`
	Bedrock   BedrockConfig     `toml:"bedrock"`
}

type AnthropicConfig struct {
	APIKey string `toml:"api_key"`
}

type BedrockConfig struct {
	Region string `toml:"region"`
	Model  string `toml:"model"`
}

type AgentConfig struct {
	MaxTurns            int  `toml:"max_turns"`
	MaxTokens           int  `toml:"max_tokens"`
	Debug               bool `toml:"debug"`
	WaitForInputTimeout int  `toml:"wait_for_input_timeout"` // seconds; default 604800 (7 days)
}

type PermissionsConfig struct {
	Bash    string `toml:"bash"`
	Write   string `toml:"write"`
	Edit    string `toml:"edit"`
	Read    string `toml:"read"`
	Glob    string `toml:"glob"`
	Grep    string `toml:"grep"`
	Default string `toml:"default"`
}

type HooksConfig struct {
	PreToolUse  []HookEntry `toml:"pre_tool_use"`
	PostToolUse []HookEntry `toml:"post_tool_use"`
}

type HookEntry struct {
	Tool    string `toml:"tool"`
	Command string `toml:"command"`
}

// Defaults returns a Config with sane default values.
func Defaults() *Config {
	return &Config{
		Provider: ProviderConfig{
			Name:  "anthropic",
			Model: "claude-sonnet-4-6",
			Bedrock: BedrockConfig{
				Region: "us-east-1",
				Model:  "us.anthropic.claude-sonnet-4-6-20251101-v1:0",
			},
		},
		Agent: AgentConfig{
			MaxTurns:            30,
			MaxTokens:           8192,
			Debug:               false,
			WaitForInputTimeout: 7 * 24 * 3600, // 7 days in seconds
		},
		Server: ServerConfig{
			Port:               8080,
			LLMConcurrency:     10,
			PerUserConcurrency: 3,
			QueueBuffer:        500,
		},
		Permissions: PermissionsConfig{
			Bash:    "auto",
			Write:   "auto",
			Edit:    "auto",
			Read:    "auto",
			Glob:    "auto",
			Grep:    "auto",
			Default: "auto",
		},
	}
}
