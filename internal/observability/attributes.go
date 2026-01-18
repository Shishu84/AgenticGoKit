package observability

import "go.opentelemetry.io/otel/attribute"

const (
	// Agent attributes
	AttrAgentName        = "agk.agent.name"
	AttrAgentType        = "agk.agent.type"
	AttrSystemPromptHash = "agk.agent.system_prompt_hash"

	// LLM attributes
	AttrLLMProvider    = "agk.llm.provider"
	AttrLLMModel       = "agk.llm.model"
	AttrLLMTemperature = "agk.llm.temperature"
	AttrLLMMaxTokens   = "agk.llm.max_tokens"
	AttrLLMTokensIn    = "agk.llm.tokens.input"
	AttrLLMTokensOut   = "agk.llm.tokens.output"
	AttrLLMCostUSD     = "agk.llm.cost.usd"
	AttrLLMRetryCount  = "agk.llm.retry_count"
	AttrLLMLatencyMs   = "agk.llm.latency_ms"

	// Tool attributes
	AttrToolName      = "agk.tool.name"
	AttrToolLatencyMs = "agk.tool.latency_ms"
	AttrMCPServer     = "agk.mcp.server"

	// Memory attributes
	AttrMemoryOperation = "agk.memory.operation"
	AttrMemorySize      = "agk.memory.size_bytes"

	// Run attributes
	AttrRunID       = "agk.run.id"
	AttrProjectName = "agk.project.name"
	AttrTemplate    = "agk.template.type"
)

const (
	EventStreamDelta = "agk.stream.delta"
	EventStreamDone  = "agk.stream.done"
	EventError       = "agk.error"
	EventRetry       = "agk.retry"
)

func LLMAttributes(provider, model string, temp float64, maxTokens int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(AttrLLMProvider, provider),
		attribute.String(AttrLLMModel, model),
		attribute.Float64(AttrLLMTemperature, temp),
		attribute.Int(AttrLLMMaxTokens, maxTokens),
	}
}

func AgentAttributes(name, agentType string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(AttrAgentName, name),
		attribute.String(AttrAgentType, agentType),
	}
}

func ToolAttributes(name string, latencyMs int64) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String(AttrToolName, name),
		attribute.Int64(AttrToolLatencyMs, latencyMs),
	}
}
