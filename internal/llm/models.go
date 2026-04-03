package llm

// ModelInfo describes a single model available for selection.
type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Default bool   `json:"default,omitempty"`
}

// modelPricing maps model IDs to USD cost per 1M tokens.
var modelPricing = map[string][2]float64{
	// [inputPer1M, outputPer1M]
	"claude-opus-4-6":   {15.0, 75.0},
	"claude-sonnet-4-6": {3.0, 15.0},
	"claude-haiku-4-5":  {0.8, 4.0},
	// Bedrock model IDs
	"us.anthropic.claude-opus-4-6-20251101-v1:0":   {15.0, 75.0},
	"us.anthropic.claude-sonnet-4-6-20251101-v1:0": {3.0, 15.0},
	"us.anthropic.claude-haiku-4-5-20251001-v1:0":  {0.8, 4.0},
}

// CalculateCostUSD returns the approximate USD cost for a given model and token counts.
// Returns 0 if the model is not in the pricing table.
func CalculateCostUSD(modelID string, inputTokens, outputTokens int) float64 {
	p, ok := modelPricing[modelID]
	if !ok {
		return 0
	}
	return (float64(inputTokens)*p[0] + float64(outputTokens)*p[1]) / 1_000_000
}

var anthropicModels = []ModelInfo{
	{ID: "claude-opus-4-6", Name: "Claude Opus 4.6"},
	{ID: "claude-sonnet-4-6", Name: "Claude Sonnet 4.6"},
	{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5"},
}

var bedrockModels = []ModelInfo{
	{ID: "us.anthropic.claude-opus-4-6-20251101-v1:0", Name: "Claude Opus 4.6 (Bedrock)"},
	{ID: "us.anthropic.claude-sonnet-4-6-20251101-v1:0", Name: "Claude Sonnet 4.6 (Bedrock)"},
	{ID: "us.anthropic.claude-haiku-4-5-20251001-v1:0", Name: "Claude Haiku 4.5 (Bedrock)"},
}

// ModelsForProvider returns the model list for the given provider name,
// marking the activeModel as default. Falls back to anthropic if unknown.
func ModelsForProvider(providerName, activeModel string) []ModelInfo {
	var base []ModelInfo
	switch providerName {
	case "bedrock":
		base = bedrockModels
	default:
		base = anthropicModels
	}

	out := make([]ModelInfo, len(base))
	copy(out, base)
	for i := range out {
		out[i].Default = out[i].ID == activeModel
	}
	return out
}
