# Native Tool Calling Implementation Guide

This document outlines all changes made to support native tool calling (function calling) across the AgenticGoKit stack, enabling LLM drivers to make structured tool calls without prompt-based parsing.

## Overview

Native tool calling allows LLM models to return structured tool invocation requests that the agent can directly execute, rather than relying on text parsing. This improves reliability and speed compared to prompt-based tool use.

### Key Benefits
- **Reliability**: Structured tool calls are less prone to parsing failures than text patterns
- **Performance**: Avoids token overhead of text formatting and parsing
- **Clarity**: Intent is explicit in the model response
- **Native Support**: Leverages modern LLM APIs (OpenAI function calling, Ollama tools, etc.)

---

## Architecture Changes

### 1. Core Type Definitions (`core/llm.go`)

Added new types to represent native tool calling contracts:

#### Tool Definition Types
```go
// ToolDefinition represents a tool available to the LLM
type ToolDefinition struct {
    Type     string                 // "function"
    Function FunctionDefinition
}

// FunctionDefinition describes a callable function
type FunctionDefinition struct {
    Name        string                 // e.g., "check_weather"
    Description string                 // e.g., "Return weather forecast for a location"
    Parameters  map[string]interface{} // JSON Schema for arguments
}
```

#### Tool Call Response Types
```go
// ToolCallResponse represents a tool call returned by the LLM
type ToolCallResponse struct {
    Type     string                  // "function"
    Function FunctionCallResponse
}

// FunctionCallResponse contains the tool name and arguments
type FunctionCallResponse struct {
    Name      string                 // Tool name to invoke
    Arguments map[string]interface{} // Parsed arguments
}
```

#### Prompt and Response Extensions
```go
// In Prompt struct:
Tools []ToolDefinition  // Tools available to the model

// In Response struct:
ToolCalls []ToolCallResponse  // Tool calls returned by the model
```

### 2. LLM Adapter Updates

#### Ollama Adapter (`internal/llm/ollama_adapter.go`)

**Non-Streaming (Call method)**:
1. **Request Building**: Include `tools` array in request body
   ```go
   if len(prompt.Tools) > 0 {
       tools := make([]map[string]interface{}, len(prompt.Tools))
       for i, tool := range prompt.Tools {
           tools[i] = map[string]interface{}{
               "type":     tool.Type,
               "function": tool.Function,
           }
       }
       requestBody["tools"] = tools
       requestBody["tool_choice"] = "auto"  // Hint model to use tools when appropriate
   }
   ```

2. **Response Parsing**: Extract `tool_calls` from response
   ```go
   var apiResp struct {
       Message struct {
           Content   string
           ToolCalls []struct {
               Function struct {
                   Name      string
                   Arguments map[string]interface{}
               }
           }
       }
   }
   
   if len(apiResp.Message.ToolCalls) > 0 {
       response.ToolCalls = make([]ToolCallResponse, len(apiResp.Message.ToolCalls))
       // Map to core types
   }
   ```

**Streaming (Stream method)**:
- **Current Limitation**: Uses text-only `/api/generate` endpoint
- **Future Work**: Upgrade to `/api/chat` endpoint with `stream=true` to emit native tool_call events mid-stream
- **Workaround**: Tool calls are parsed from final text content post-stream

#### Wrapper Updates (`internal/llm/wrappers.go`)

Public wrapper forwards tool definitions and maps tool call responses:
```go
// In Call method:
if len(internalPrompt.Tools) > 0 {
    publicPrompt.Tools = convertToolsToPublic(internalPrompt.Tools)
}

response := a.adapter.Call(ctx, publicPrompt)
if len(response.ToolCalls) > 0 {
    internalResponse.ToolCalls = convertToolCallsToInternal(response.ToolCalls)
}
```

---

## Agent Runtime Changes (`v1beta/agent_impl.go`)

### Tool Schema Extraction

```go
// convertToolsToLLMFormat converts v1beta tools to llm.ToolDefinition
func convertToolsToLLMFormat(tools []Tool) []llm.ToolDefinition {
    for _, tool := range tools {
        res[i] = llm.ToolDefinition{
            Type: "function",
            Function: llm.FunctionDefinition{
                Name:        tool.Name(),
                Description: tool.Description(),
                Parameters:  getToolSchema(tool),
            },
        }
    }
    return res
}

// getToolSchema returns JSON Schema, with fallback to minimal schema
func getToolSchema(tool Tool) map[string]interface{} {
    if ts, ok := tool.(ToolWithSchema); ok {
        return ts.JSONSchema()  // Custom schema from tool
    }
    // Fallback minimal schema
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "input": map[string]interface{}{
                "type": "string",
                "description": "User input",
            },
        },
    }
}
```

### Tool Call Execution

#### Native Tool Calls (structured)
```go
func (a *realAgent) executeNativeToolsAndContinue(
    ctx context.Context,
    initialResponse llm.Response,
    originalPrompt llm.Prompt,
    maxIterations int,
) (string, []ToolCall, error) {
    // 1. Parse Response.ToolCalls (structured calls from LLM)
    // 2. Deduplicate using buildToolSignature (avoid repeated calls)
    // 3. Execute each call via a.executeTool()
    // 4. Feed results back to LLM in continuation prompt
    // 5. Loop until no more tool calls or max iterations
}
```

Key features:
- **Deduplication**: Track executed calls by signature to avoid repeats
- **Iteration Limit**: Single-tool runs limited to 1 iteration by default
- **Result Feeding**: Tool results appended to next prompt for reasoning

#### Text-Parsed Tool Calls (fallback)
```go
func (a *realAgent) executeToolsAndContinue(
    ctx context.Context,
    initialResponse string,
    originalPrompt llm.Prompt,
    maxIterations int,
) (string, []ToolCall, error) {
    // 1. Parse tool calls from text using ParseToolCalls()
    // 2. Apply single-call policy if needed
    // 3. Execute and feed back results
}
```

### Single-Call Policy

When multiple tool calls are detected and only one tool is enabled:

```go
type SingleCallPolicy string
// Values: "best" (default), "first", "all"

func selectBestToolCallFromInput(input string, calls []ToolCall) ToolCall {
    // Match arguments to user input keywords
    // Prefer calls with arguments appearing in the input
    // Handle common abbreviations (sf→San Francisco)
}
```

**Config Integration** (`v1beta/config.go`):
```go
type ToolsConfig struct {
    Enabled          bool
    SingleCallPolicy string  // "best", "first", or "all"
    // ... other fields
}
```

### Result Synthesis

When LLM returns empty content but tools executed:

```go
func formatToolCallsAsContent(calls []ToolCall) string {
    // Single tool: return forecast directly
    // Multiple tools: build compact list summary
    // Returns user-friendly string for Result.Content
}
```

**Integration in Run()**:
```go
// After tool execution:
if len(toolCalls) > 0 {
    summary := formatToolCallsAsContent(toolCalls)
    if summary != "" {
        finalResponse = summary  // Use tool results as response
    }
}
```

---

## Configuration and Discovery (`v1beta/`)

### Tool Registration (`tool_discovery.go`)

```go
// RegisterInternalTool registers a tool for automatic discovery
vnext.RegisterInternalTool("check_weather", func() vnext.Tool {
    return &WeatherTool{}
})

// DiscoverInternalTools returns all registered tools
func DiscoverInternalTools() ([]Tool, error) {
    // Iterate registry, instantiate each tool
}
```

### Default Configuration (`tools.go`)

```go
func DefaultToolsConfig() *ToolsConfig {
    return &ToolsConfig{
        Enabled:          true,
        MaxRetries:       3,
        Timeout:          30 * time.Second,
        SingleCallPolicy: "best",  // Default policy
        // ... other defaults
    }
}
```

---

## Example: LangChain-Style Weather Demo

Location: `examples/langchain-weather-demo/main.go`

### Tool Definition with Schema

```go
type WeatherTool struct{}

func (t *WeatherTool) Name() string { return "check_weather" }
func (t *WeatherTool) Description() string {
    return "Return the weather forecast for the specified location"
}

// JSONSchema provides schema for native tool calling
func (t *WeatherTool) JSONSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type":        "string",
                "description": "City name or abbreviation",
                "examples":    []string{"San Francisco", "sf", "Tokyo"},
            },
        },
        "required": []string{"location"},
    }
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
    location := args["location"].(string)
    forecast := fmt.Sprintf("It's always sunny in %s ☀️", location)
    return &ToolResult{Success: true, Content: map[string]interface{}{"forecast": forecast}}, nil
}
```

### Registration and Agent Setup

```go
// Automatic registration (called at init)
vnext.RegisterInternalTool("check_weather", func() vnext.Tool {
    return &WeatherTool{}
})

// Agent config enables tools
agent, _ := vnext.NewBuilder("weather-agent").
    WithConfig(&vnext.Config{
        Name:         "weather-agent",
        SystemPrompt: "You are a helpful assistant.",
        LLM: vnext.LLMConfig{
            Provider:    "ollama",
            Model:       "functiongemma:latest",
            Temperature: 0.0,
            MaxTokens:   150,
        },
        Tools: &vnext.ToolsConfig{
            Enabled:          true,
            SingleCallPolicy: "best",  // Single best tool call per run
        },
        Memory: &vnext.MemoryConfig{Enabled: false},
        Timeout: 30 * time.Second,
    }).
    WithPreset(vnext.ChatAgent).
    Build()

// Simple execution
res, _ := agent.Run(ctx, "what is the weather in sf")
fmt.Println(res.Content)  // "It's always sunny in sf ☀️"
```

---

## How to Add Native Tool Calling to Other LLM Drivers

### Step 1: Update the LLM Adapter

**File**: `internal/llm/{provider}_adapter.go`

Implement tool support in the `Call()` method:

```go
func (a *YourAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
    // 1. Build request with tools
    requestBody := map[string]interface{}{
        "model":   a.model,
        "messages": messages,
        // ... other fields
    }
    
    // 2. Include tools if provided
    if len(prompt.Tools) > 0 {
        tools := convertToolsToProviderFormat(prompt.Tools)
        requestBody["tools"] = tools
        requestBody["tool_choice"] = "auto"  // or provider's equivalent
    }
    
    // 3. Parse response
    var apiResp struct {
        Message struct {
            Content   string
            ToolCalls []ProviderToolCall  // Provider-specific structure
        }
    }
    json.NewDecoder(resp.Body).Decode(&apiResp)
    
    // 4. Map to core types
    response := Response{
        Content: apiResp.Message.Content,
    }
    if len(apiResp.Message.ToolCalls) > 0 {
        response.ToolCalls = mapToolCalls(apiResp.Message.ToolCalls)
    }
    return response, nil
}

// Helper: Convert core.ToolDefinition to provider format
func convertToolsToProviderFormat(tools []llm.ToolDefinition) interface{} {
    // Provider-specific conversion (OpenAI format, Anthropic format, etc.)
}

// Helper: Map provider tool calls to core.ToolCallResponse
func mapToolCalls(calls []ProviderToolCall) []llm.ToolCallResponse {
    // Extract name and arguments, return core types
}
```

### Step 2: Handle Streaming (Optional but Recommended)

**File**: `internal/llm/{provider}_adapter.go` - `Stream()` method

If the provider supports streaming with tool calls:

```go
func (a *YourAdapter) Stream(ctx context.Context, prompt Prompt) (<-chan Token, error) {
    // Similar to Call(), but stream responses
    // Include tools in request if available
    // Parse tool_call events from stream
    
    // Note: Current implementation may need event parsing for delta tool calls
}
```

### Step 3: Update the Plugin Wrapper (if applicable)

**File**: `plugins/llm/{provider}/{provider}.go`

Forward tool definitions through public wrapper:

```go
func (a *PublicAdapter) Call(ctx context.Context, prompt core.Prompt) (core.Response, error) {
    // Convert to internal types
    internalPrompt := core.Prompt{
        System: prompt.System,
        User:   prompt.User,
    }
    
    // Forward tools if present
    if len(prompt.Tools) > 0 {
        internalPrompt.Tools = convertToInternalTools(prompt.Tools)
    }
    
    internalResp, err := a.adapter.Call(ctx, internalPrompt)
    
    // Map tool calls back to public types
    if len(internalResp.ToolCalls) > 0 {
        resp.ToolCalls = convertToPublicToolCalls(internalResp.ToolCalls)
    }
    return resp, nil
}
```

### Step 4: Testing

Create a test similar to Ollama tests:

```go
func TestYourAdapterWithTools(t *testing.T) {
    adapter := NewYourAdapter(apiKey, model)
    
    prompt := Prompt{
        System: "You are a helpful assistant.",
        User:   "What tool would you use?",
        Tools: []ToolDefinition{
            {
                Type: "function",
                Function: FunctionDefinition{
                    Name:        "test_tool",
                    Description: "A test tool",
                    Parameters: map[string]interface{}{
                        "type": "object",
                        "properties": map[string]interface{}{
                            "arg": map[string]interface{}{
                                "type": "string",
                            },
                        },
                    },
                },
            },
        },
    }
    
    resp, err := adapter.Call(context.Background(), prompt)
    assert.NoError(t, err)
    assert.NotEmpty(t, resp.ToolCalls)
}
```

---

## Provider-Specific Notes

### OpenAI / Azure OpenAI
- Tools are called `functions` in the API
- Response structure: `message.tool_calls[].function.{name, arguments}`
- Already implemented: See `internal/llm/openai_adapter.go`

### Ollama
- Uses `tools` array in `/api/chat` endpoint
- Response: `message.tool_calls[].function.{name, arguments}`
- **Current Implementation**: Non-streaming only
- **Future Enhancement**: Upgrade streaming to chat endpoint for native tool_call events

### Anthropic (Claude)
- Uses `tools` array in request
- Response: `content[].type == "tool_use"` blocks
- Requires message reconstruction with `tool_results` role

### HuggingFace / vLLM
- May require custom prompt formatting for tool definitions
- Check provider documentation for tool calling support
- Some models use text-based tool calling (fallback to ParseToolCalls)

---

## Migration Path

For existing LLM drivers without native tool support:

1. **No Action Required**: Text-parsed tool calling still works via `ParseToolCalls()`
2. **Gradual Enhancement**: Add native support when convenient
3. **Backward Compatibility**: Both native and text-parsed paths coexist

The agent automatically prefers native tool calls (`Response.ToolCalls`) when available, falling back to text parsing.

---

## Performance Considerations

### Why Native Tool Calling is Faster
- **No Token Overhead**: Tool calls don't consume tokens for formatting/parsing instructions
- **Direct Execution**: Structured format prevents parsing ambiguities
- **Reduced Rounds**: Clearer intent can reduce LLM iterations

### Benchmarking
When benchmarking, ensure both implementations:
- Call the same LLM endpoint
- Use the same model
- Warm the model cache first
- Include full LLM inference time (not just tool execution)

Example:
```bash
# Go example with native tools
time go run examples/langchain-weather-demo/main.go

# Python equivalent should call same Ollama model
time python examples/langchain-weather-demo/weather_tool.py
```

Expected result: LLM latency dominates (seconds), not tool registration or execution (milliseconds).

---

## Files Modified Summary

| File | Change | Purpose |
|------|--------|---------|
| `core/llm.go` | Added tool types | Define native tool calling contract |
| `internal/llm/ollama_adapter.go` | Updated Call() | Send tools, parse tool_calls |
| `internal/llm/wrappers.go` | Updated Call() | Forward tool definitions |
| `v1beta/config.go` | Added SingleCallPolicy | Configure tool execution policy |
| `v1beta/tools.go` | Set defaults | Default policy to "best" |
| `v1beta/agent_impl.go` | Multiple methods | Execute native & text-parsed calls, synthesize results |
| `v1beta/tool_discovery.go` | Added helpers | Tool registration and discovery |
| `examples/langchain-weather-demo/main.go` | New example | Demonstrate minimal tool-using agent |

---

## Conclusion

This implementation provides a flexible, extensible native tool calling framework that:
- Works with modern LLM APIs that support function calling
- Falls back gracefully to text parsing for others
- Gives users explicit control via `SingleCallPolicy`
- Scales from simple single-tool examples to complex multi-tool workflows

For questions or contributions, refer to the relevant adapter files and test cases.
