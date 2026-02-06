package ollama

import (
	"context"

	"github.com/agenticgokit/agenticgokit/core"
	"github.com/agenticgokit/agenticgokit/internal/llm"
)

type providerAdapter struct{ adapter *llm.PublicProviderAdapter }

func (a *providerAdapter) Call(ctx context.Context, p core.Prompt) (core.Response, error) {
	ip := llm.PublicPrompt{
		System: p.System,
		User:   p.User,
		Parameters: llm.PublicModelParameters{
			Temperature: p.Parameters.Temperature,
			MaxTokens:   p.Parameters.MaxTokens,
		},
		// Forward native tool definitions to the underlying provider
		Tools: convertCoreToolsToInternal(p.Tools),
	}
	resp, err := a.adapter.Call(ctx, ip)
	if err != nil {
		return core.Response{}, err
	}
	return core.Response{
		Content: resp.Content,
		Usage: core.UsageStats{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
		FinishReason: resp.FinishReason,
		// Surface structured tool calls back to core
		ToolCalls: convertInternalToolCallsToCore(resp.ToolCalls),
	}, nil
}
func (a *providerAdapter) Stream(ctx context.Context, p core.Prompt) (<-chan core.Token, error) {
	ip := llm.PublicPrompt{
		System: p.System,
		User:   p.User,
		Parameters: llm.PublicModelParameters{
			Temperature: p.Parameters.Temperature,
			MaxTokens:   p.Parameters.MaxTokens,
		},
		// Forward native tool definitions to the underlying provider
		Tools: convertCoreToolsToInternal(p.Tools),
	}
	ich, err := a.adapter.Stream(ctx, ip)
	if err != nil {
		return nil, err
	}
	och := make(chan core.Token)
	go func() {
		defer close(och)
		for t := range ich {
			och <- core.Token{Content: t.Content, Error: t.Error}
		}
	}()
	return och, nil
}
func (a *providerAdapter) Embeddings(ctx context.Context, texts []string) ([][]float64, error) {
	return a.adapter.Embeddings(ctx, texts)
}

func factory(cfg core.LLMProviderConfig) (core.ModelProvider, error) {
	wrapper, err := llm.NewOllamaAdapterWrapped(cfg.BaseURL, cfg.Model, cfg.MaxTokens, float32(cfg.Temperature), cfg.HTTPTimeout)
	if err != nil {
		return nil, err
	}
	return &providerAdapter{adapter: llm.NewPublicProviderAdapter(wrapper)}, nil
}

func init() { core.RegisterModelProviderFactory("ollama", factory) }

// convertCoreToolsToInternal maps core.ToolDefinition to llm.ToolDefinition for passing through wrappers.
func convertCoreToolsToInternal(tools []core.ToolDefinition) []llm.ToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	res := make([]llm.ToolDefinition, len(tools))
	for i, t := range tools {
		res[i] = llm.ToolDefinition{
			Type: t.Type,
			Function: llm.FunctionDefinition{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			},
		}
	}
	return res
}

// convertInternalToolCallsToCore maps llm.ToolCallResponse to core.ToolCallResponse for returning to callers.
func convertInternalToolCallsToCore(calls []llm.ToolCallResponse) []core.ToolCallResponse {
	if len(calls) == 0 {
		return nil
	}
	res := make([]core.ToolCallResponse, len(calls))
	for i, c := range calls {
		res[i] = core.ToolCallResponse{
			ID:   c.ID,
			Type: c.Type,
			Function: core.FunctionCallResponse{
				Name:      c.Function.Name,
				Arguments: c.Function.Arguments,
			},
		}
	}
	return res
}
