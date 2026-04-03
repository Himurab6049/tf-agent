package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

// BedrockProvider implements Provider using AWS Bedrock.
type BedrockProvider struct {
	client *bedrockruntime.Client
	model  string
	debug  bool
}

// NewBedrockProvider creates an AWS Bedrock provider.
// When debug is true, raw requests are logged to stderr.
func NewBedrockProvider(region, model string, debug bool) (*BedrockProvider, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	return &BedrockProvider{
		client: bedrockruntime.NewFromConfig(cfg),
		model:  model,
		debug:  debug,
	}, nil
}

func (p *BedrockProvider) Name() string { return "bedrock" }

// bedrockRequest mirrors the Anthropic request body shape.
type bedrockRequest struct {
	AnthropicVersion string             `json:"anthropic_version"`
	MaxTokens        int                `json:"max_tokens"`
	System           string             `json:"system,omitempty"`
	Messages         []anthropicMessage `json:"messages"`
	Tools            []ToolSchema       `json:"tools,omitempty"`
}

// Stream sends the request via Bedrock's InvokeModelWithResponseStream.
func (p *BedrockProvider) Stream(ctx context.Context, req Request) (<-chan Event, error) {
	ch := make(chan Event, 64)
	go func() {
		defer close(ch)
		if err := p.doStream(ctx, req, ch); err != nil {
			ch <- Event{Type: EventError, Err: err}
		}
	}()
	return ch, nil
}

func (p *BedrockProvider) doStream(ctx context.Context, req Request, ch chan<- Event) error {
	br := bedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        req.MaxTokens,
		System:           req.System,
		Messages:         convertMessages(req.Messages),
		Tools:            req.Tools,
	}
	if len(br.Tools) == 0 {
		br.Tools = nil
	}

	payload, err := json.Marshal(br)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if p.debug {
		log.Printf("[DEBUG] bedrock request model=%s body=%s", p.model, payload)
	}

	input := &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(p.model),
		Body:        payload,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	}

	resp, err := p.client.InvokeModelWithResponseStream(ctx, input)
	if err != nil {
		return fmt.Errorf("invoke bedrock: %w", err)
	}

	stream := resp.GetStream()
	defer stream.Close()

	// Bedrock wraps Anthropic chunks in an event envelope. Each chunk's bytes
	// are the same JSON as Anthropic's non-streaming response delta events.
	// We parse them the same way using a lightweight inline decoder.
	var tool *pendingTool

	for event := range stream.Events() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var wrapper struct {
				Type string `json:"type"`
				// We reuse the same SSE event fields.
				ContentBlock *struct {
					Type string `json:"type"`
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"content_block"`
				Delta *struct {
					Type        string `json:"type"`
					Text        string `json:"text"`
					PartialJSON string `json:"partial_json"`
					StopReason  string `json:"stop_reason"`
				} `json:"delta"`
				Message *struct {
					Usage struct {
						InputTokens  int `json:"input_tokens"`
						CacheRead    int `json:"cache_read_input_tokens"`
						CacheCreated int `json:"cache_creation_input_tokens"`
					} `json:"usage"`
				} `json:"message"`
				Usage *struct {
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
			}

			if err := json.Unmarshal(v.Value.Bytes, &wrapper); err != nil {
				continue
			}

			switch wrapper.Type {
			case "message_start":
				if wrapper.Message != nil {
					ch <- Event{Type: EventUsage, Usage: &UsageEvent{
						InputTokens:  wrapper.Message.Usage.InputTokens,
						CacheRead:    wrapper.Message.Usage.CacheRead,
						CacheCreated: wrapper.Message.Usage.CacheCreated,
					}}
				}

			case "content_block_start":
				if wrapper.ContentBlock != nil && wrapper.ContentBlock.Type == "tool_use" {
					tool = &pendingTool{
						id:   wrapper.ContentBlock.ID,
						name: wrapper.ContentBlock.Name,
					}
				}

			case "content_block_delta":
				if wrapper.Delta != nil {
					switch wrapper.Delta.Type {
					case "text_delta":
						ch <- Event{Type: EventText, Delta: wrapper.Delta.Text}
					case "input_json_delta":
						if tool != nil {
							tool.inputJSON.WriteString(wrapper.Delta.PartialJSON)
						}
					}
				}

			case "content_block_stop":
				if tool != nil {
					s := tool.inputJSON.String()
					if s == "" {
						s = "{}"
					}
					ch <- Event{
						Type: EventToolUse,
						ToolUse: &ToolUseEvent{
							ID:    tool.id,
							Name:  tool.name,
							Input: json.RawMessage(s),
						},
					}
					tool = nil
				}

			case "message_delta":
				if wrapper.Usage != nil {
					ch <- Event{Type: EventUsage, Usage: &UsageEvent{OutputTokens: wrapper.Usage.OutputTokens}}
				}
				if wrapper.Delta != nil && wrapper.Delta.StopReason != "" {
					ch <- Event{Type: EventStop, StopReason: wrapper.Delta.StopReason}
				}

			case "message_stop":
				ch <- Event{Type: EventStop, StopReason: "end_turn"}
			}
		}
	}
	return stream.Err()
}
