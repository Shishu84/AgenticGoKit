# Native Tool Calling Support Analysis for AgenticGoKit

**Date:** January 7, 2026  
**Author:** Analysis of AgenticGoKit Architecture  
**Status:** Analysis & Roadmap

## Executive Summary

AgenticGoKit currently implements **text-based tool calling** where tools are described in prompts and the LLM must emit function calls as text. Modern models like `functiongemma:latest` support **native tool calling** where tools are sent as structured data to the API, and the model returns structured tool call objects.

This document analyzes what changes are needed to support native tool calling, similar to how LangChain uses Ollama's native API.

---

## 1. Current Architecture: Text-Based Tool Calling

### How It Works Now

```
┌─────────────┐
│ Agent       │
│ - tools []  │
└──────┬──────┘
       │
       │ 1. FormatToolsForPrompt() converts tools to text
       ▼
┌──────────────────────────────────────────┐
│ System Prompt (Enhanced)                 │
│                                          │
│ "You have access to:                     │
│  - check_weather: Get weather forecast   │
│                                          │
│ To use: tool_name(arg=\"value\")"        │
└──────┬───────────────────────────────────┘
       │
       │ 2. Call LLM with text prompt
       ▼
┌──────────────────┐
│ LLM Provider     │
│ (Ollama)         │
└──────┬───────────┘
       │
       │ 3. Returns text response
       ▼
"check_weather(location=\"Tokyo\")"
       │
       │ 4. ParseToolCalls() extracts tool calls from text
       ▼
┌──────────────────┐
│ Execute Tool     │
└──────────────────┘
```

### Current Flow in Code

1. **Tool Registration** (`v1beta/tool_discovery.go`)
   - Tools registered via `RegisterInternalTool()`
   - Stored in global registry

2. **Prompt Enhancement** (`v1beta/agent_impl.go:281`)
   ```go
   toolDescriptions := FormatToolsForPrompt(a.tools)
   prompt.System = prompt.System + toolDescriptions
   ```

3. **LLM Call** (`internal/llm/ollama_adapter.go:145`)
   ```go
   requestBody := map[string]interface{}{
       "model":       o.model,
       "messages":    messages,
       "max_tokens":  finalMaxTokens,
       "temperature": finalTemperature,
       "stream":      false,
       // NO "tools" PARAMETER!
   }
   ```

4. **Parse Response** (`v1beta/utils.go:499`)
   ```go
   func ParseToolCalls(content string) []ToolCall {
       // Parses text: "check_weather(location=\"Tokyo\")"
       // Or ReAct format: "Action: check_weather\nAction Input: {\"location\":\"Tokyo\"}"
   }
   ```

### Problems with Text-Based Approach

❌ **Model Dependent**: Small models (gemma3:1b) ignore instructions and hallucinate  
❌ **Unreliable Parsing**: Text parsing can fail on format variations  
❌ **No Guarantees**: Model may not call tools even when instructed  
❌ **Token Waste**: Tool descriptions consume prompt tokens  
❌ **Slower**: Text parsing + multi-turn LLM calls for tool execution

---

## 2. Native Tool Calling: How It Should Work

### Ollama Native API

From research and LangChain implementation:

```json
POST /api/chat
{
  "model": "functiongemma",
  "messages": [
    {"role": "user", "content": "What's the weather in Tokyo?"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "check_weather",
        "description": "Get weather forecast for a location",
        "parameters": {
          "type": "object",
          "properties": {
            "location": {
              "type": "string",
              "description": "City name"
            }
          },
          "required": ["location"]
        }
      }
    }
  ]
}
```

**Response:**
```json
{
  "message": {
    "role": "assistant",
    "content": "",
    "tool_calls": [
      {
        "function": {
          "name": "check_weather",
          "arguments": {
            "location": "Tokyo"
          }
        }
      }
    ]
  }
}
```

### Advantages of Native Tool Calling

✅ **Guaranteed Execution**: Model trained to use tools, not ignore them  
✅ **Structured Output**: Parsed JSON, no text parsing needed  
✅ **Type Safety**: Arguments validated by model  
✅ **Faster**: Single LLM call returns structured tool call  
✅ **Model Optimized**: Models like `functiongemma` designed for this

---

## 3. Required Changes

### 3.1 Core Type Definitions

**File:** `core/llm.go` and `internal/llm/types.go`

```go
// Add to Prompt struct
type Prompt struct {
    System     string
    User       string
    Images     []ImageData
    Audio      []AudioData
    Video      []VideoData
    Parameters ModelParameters
    Tools      []ToolDefinition  // NEW: Native tool definitions
}

// NEW: OpenAI-compatible tool definition
type ToolDefinition struct {
    Type     string            `json:"type"`     // "function"
    Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
}

// Add to Response struct
type Response struct {
    Content      string
    Usage        UsageStats
    FinishReason string
    ToolCalls    []ToolCallResponse  // NEW: Structured tool calls from model
    // ... existing fields
}

// NEW: Structured tool call from LLM response
type ToolCallResponse struct {
    ID       string                 `json:"id,omitempty"`
    Type     string                 `json:"type"`     // "function"
    Function FunctionCallResponse   `json:"function"`
}

type FunctionCallResponse struct {
    Name      string                 `json:"name"`
    Arguments map[string]interface{} `json:"arguments"` // Parsed JSON
}
```

### 3.2 Ollama Adapter Changes

**File:** `internal/llm/ollama_adapter.go`

```go
// In Call() method (around line 145)
func (o *OllamaAdapter) Call(ctx context.Context, prompt Prompt) (Response, error) {
    // ... existing code ...

    requestBody := map[string]interface{}{
        "model":       o.model,
        "messages":    messages,
        "max_tokens":  finalMaxTokens,
        "temperature": finalTemperature,
        "stream":      false,
    }

    // NEW: Add tools if provided
    if len(prompt.Tools) > 0 {
        tools := make([]map[string]interface{}, len(prompt.Tools))
        for i, tool := range prompt.Tools {
            tools[i] = map[string]interface{}{
                "type":     tool.Type,
                "function": tool.Function,
            }
        }
        requestBody["tools"] = tools
    }

    // ... existing request code ...

    // NEW: Parse tool calls from response
    var apiResp struct {
        Message struct {
            Content   string `json:"content"`
            ToolCalls []struct {
                Function struct {
                    Name      string                 `json:"name"`
                    Arguments map[string]interface{} `json:"arguments"`
                } `json:"function"`
            } `json:"tool_calls,omitempty"`
        } `json:"message"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
        return Response{}, fmt.Errorf("failed to decode response: %w", err)
    }

    response := Response{
        Content: apiResp.Message.Content,
    }

    // NEW: Convert tool calls to our format
    if len(apiResp.Message.ToolCalls) > 0 {
        response.ToolCalls = make([]ToolCallResponse, len(apiResp.Message.ToolCalls))
        for i, tc := range apiResp.Message.ToolCalls {
            response.ToolCalls[i] = ToolCallResponse{
                Type: "function",
                Function: FunctionCallResponse{
                    Name:      tc.Function.Name,
                    Arguments: tc.Function.Arguments,
                },
            }
        }
    }

    return response, nil
}
```

### 3.3 v1beta Agent Changes

**File:** `v1beta/agent_impl.go`

```go
// In execute() method (around line 270)
func (a *realAgent) execute(ctx context.Context, input string, opts *RunOptions) (*Result, error) {
    // ... existing code ...

    prompt := llm.Prompt{
        System: a.config.SystemPrompt,
        User:   input,
    }

    // CHANGED: Instead of adding tools to system prompt as text,
    // convert v1beta.Tool to llm.ToolDefinition
    if len(a.tools) > 0 {
        prompt.Tools = convertToolsToLLMFormat(a.tools)
        
        Logger().Debug().
            Int("tool_count", len(a.tools)).
            Msg("Added native tool definitions to prompt")
    }

    // ... existing LLM call ...

    response, err := a.llmProvider.Call(ctx, prompt)
    if err != nil {
        // ...
    }

    // CHANGED: Check response.ToolCalls instead of parsing text
    if len(response.ToolCalls) > 0 {
        // Execute tools and continue
        finalResponse, toolCalls, toolErr := a.executeNativeToolsAndContinue(ctx, response, prompt, 5)
        // ...
    }

    // ... rest of method ...
}

// NEW: Convert v1beta.Tool to llm.ToolDefinition
func convertToolsToLLMFormat(tools []Tool) []llm.ToolDefinition {
    result := make([]llm.ToolDefinition, len(tools))
    for i, tool := range tools {
        // Convert tool to JSON Schema format
        // This needs tool interface extension (see below)
        result[i] = llm.ToolDefinition{
            Type: "function",
            Function: llm.FunctionDefinition{
                Name:        tool.Name(),
                Description: tool.Description(),
                Parameters:  tool.JSONSchema(), // NEW METHOD NEEDED
            },
        }
    }
    return result
}

// NEW: Execute tools from native tool calls
func (a *realAgent) executeNativeToolsAndContinue(
    ctx context.Context,
    response llm.Response,
    originalPrompt llm.Prompt,
    maxIterations int,
) (string, []ToolCall, error) {
    // Execute each tool_call from response.ToolCalls
    // Build conversation with tool results
    // Call LLM again with tool role messages
    // Continue until no more tool calls
}
```

### 3.4 Tool Interface Extension

**File:** `v1beta/agent.go`

```go
// Current Tool interface
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, args map[string]interface{}) (*ToolResult, error)
}

// EXTENDED Tool interface (backward compatible via optional interface)
type ToolWithSchema interface {
    Tool
    JSONSchema() map[string]interface{}
}

// Helper for tools that don't implement JSONSchema
func getToolSchema(tool Tool) map[string]interface{} {
    if ts, ok := tool.(ToolWithSchema); ok {
        return ts.JSONSchema()
    }
    
    // Fallback: Generate basic schema
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "args": map[string]interface{}{
                "type": "object",
                "description": "Tool arguments",
            },
        },
    }
}
```

### 3.5 Example Usage Updates

**File:** `examples/langchain-weather-demo/main.go`

```go
// Extend WeatherTool to implement ToolWithSchema
type WeatherTool struct{}

func (t *WeatherTool) Name() string { return "check_weather" }
func (t *WeatherTool) Description() string {
    return "Return the weather forecast for the specified location"
}

// NEW: Implement JSONSchema for native tool calling
func (t *WeatherTool) JSONSchema() map[string]interface{} {
    return map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "location": map[string]interface{}{
                "type":        "string",
                "description": "City name",
            },
        },
        "required": []string{"location"},
    }
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
    location, ok := args["location"].(string)
    if !ok || location == "" {
        return &vnext.ToolResult{Success: false, Error: "location parameter required"}, 
               fmt.Errorf("location required")
    }

    forecast := fmt.Sprintf("It's always sunny in %s ☀️", location)
    return &vnext.ToolResult{
        Success: true,
        Content: map[string]interface{}{
            "location": location,
            "forecast": forecast,
        },
    }, nil
}
```

---

## 4. Implementation Roadmap

### Phase 1: Core Types & Interfaces (2-3 days)

- [ ] Add `Tools []ToolDefinition` to `Prompt` struct
- [ ] Add `ToolCalls []ToolCallResponse` to `Response` struct
- [ ] Define `ToolDefinition`, `FunctionDefinition`, `ToolCallResponse` types
- [ ] Create `ToolWithSchema` optional interface
- [ ] Add backward compatibility helpers

**Files:**
- `core/llm.go`
- `internal/llm/types.go`
- `v1beta/agent.go`

### Phase 2: Ollama Adapter Native Tools (2-3 days)

- [ ] Add `tools` parameter to Ollama `/api/chat` requests
- [ ] Parse `tool_calls` from Ollama responses
- [ ] Handle tool message type in conversation history
- [ ] Add feature detection (check if model supports tools)
- [ ] Update tests

**Files:**
- `internal/llm/ollama_adapter.go`
- `internal/llm/ollama_adapter_test.go`

### Phase 3: Agent Integration (3-4 days)

- [ ] Implement `convertToolsToLLMFormat()` helper
- [ ] Add `executeNativeToolsAndContinue()` method
- [ ] Update `execute()` to use native tools when available
- [ ] Add fallback to text-based tools for unsupported models
- [ ] Implement conversation history with tool messages
- [ ] Update metrics tracking

**Files:**
- `v1beta/agent_impl.go`
- `v1beta/utils.go`

### Phase 4: Configuration & Options (1-2 days)

- [ ] Add `UseNativeToolCalling bool` to config
- [ ] Add `ToolCallingMode` enum: auto/native/text/off
- [ ] Update `RunOptions` to support tool calling preferences
- [ ] Add model capability detection

**Files:**
- `v1beta/config.go`
- `v1beta/agent.go`

### Phase 5: Examples & Documentation (2-3 days)

- [ ] Update `langchain-weather-demo` with `JSONSchema()`
- [ ] Create new example: `native-tool-calling-demo`
- [ ] Add comparison example: native vs text-based
- [ ] Update documentation
- [ ] Create migration guide

**Files:**
- `examples/langchain-weather-demo/main.go`
- `examples/native-tool-calling-demo/`
- `docs/NATIVE_TOOL_CALLING.md`
- `docs/MIGRATION.md`

### Phase 6: OpenAI & Other Providers (2-3 days)

- [ ] Add native tools to OpenAI adapter
- [ ] Verify compatibility with other providers
- [ ] Standardize tool calling across all providers
- [ ] Add provider capability matrix

**Files:**
- `internal/llm/openai_adapter.go`
- `plugins/llm/*/`

### Phase 7: Testing & Polish (2-3 days)

- [ ] Add comprehensive unit tests
- [ ] Add integration tests with real models
- [ ] Performance benchmarks: native vs text
- [ ] Update CI/CD
- [ ] Beta testing with community

**Total Estimate:** 14-21 days (~3-4 weeks)

---

## 5. Backward Compatibility

### Strategy

1. **Optional Interface**: `ToolWithSchema` is optional
   - Old tools without `JSONSchema()` still work via text-based calling
   - New tools implement `JSONSchema()` for native calling

2. **Auto-Detection**: Check if LLM supports native tools
   ```go
   if supportsNativeTools(model) && hasToolSchemas(tools) {
       // Use native calling
   } else {
       // Fallback to text-based
   }
   ```

3. **Configuration**: Allow explicit control
   ```go
   config.ToolCallingMode = "auto"  // Default: auto-detect
   config.ToolCallingMode = "native"  // Force native
   config.ToolCallingMode = "text"    // Force text-based
   ```

4. **Graceful Degradation**: If native fails, fall back to text

---

## 6. Benefits After Implementation

### For Users

✅ **Reliability**: Tools actually get called (no hallucinations)  
✅ **Performance**: Faster, fewer tokens, single LLM call  
✅ **Compatibility**: Works with modern tool-calling models  
✅ **Type Safety**: Structured arguments, less parsing errors

### For Developers

✅ **Simpler Code**: No text parsing hacks  
✅ **Better DX**: JSON Schema = clear tool contracts  
✅ **Debugging**: Structured logs of tool calls  
✅ **Testing**: Mock tool calls with structured data

---

## 7. Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Breaking changes to `Prompt`/`Response` | High | Use optional fields, maintain text-based fallback |
| Not all models support native tools | Medium | Feature detection + auto-fallback |
| JSON Schema complexity | Low | Provide helpers, examples, generators |
| Performance regression | Low | Benchmark both modes, make configurable |
| Third-party provider support | Medium | Start with Ollama, extend incrementally |

---

## 8. Success Metrics

- [ ] `functiongemma:latest` successfully calls tools 100% of the time
- [ ] Performance: <500ms for tool call (vs current 2s+)
- [ ] Token savings: 30-50% reduction in prompt tokens
- [ ] Zero text parsing errors
- [ ] Backward compatibility: All existing tests pass
- [ ] Community adoption: 5+ examples using native tools

---

## 9. Alternatives Considered

### Option A: Ollama-Only Implementation
**Pros:** Faster to ship, focused scope  
**Cons:** Not portable to other providers  
**Decision:** Rejected - need provider-agnostic solution

### Option B: Separate NativeToolAgent
**Pros:** No risk to existing code  
**Cons:** Code duplication, confusing for users  
**Decision:** Rejected - prefer unified interface

### Option C: Text + Structured Hybrid
**Pros:** Best of both worlds  
**Cons:** Complex, maintenance burden  
**Decision:** Rejected - auto-detection is cleaner

### Selected: **Auto-Detection with Fallback**
- Check model capability
- Use native if available
- Fall back to text-based
- Allow explicit override

---

## 10. Conclusion

Native tool calling is essential for modern agentic frameworks. The implementation requires:

1. **Core changes**: Add `Tools` to `Prompt`, `ToolCalls` to `Response`
2. **Adapter updates**: Ollama, OpenAI, others
3. **Agent logic**: Native tool execution loop
4. **Backward compatibility**: Optional `JSONSchema()` interface

**Estimated effort:** 3-4 weeks  
**Priority:** High (critical for competitive feature parity)  
**Next steps:** Review this analysis, approve roadmap, start Phase 1

---

## Appendix A: Code Samples

### Before (Text-Based)
```go
// System prompt
"You have access to:\n- check_weather: Get weather\nTo use: tool_name(arg=\"value\")"

// LLM response (text)
"Let me check the weather. check_weather(location=\"Tokyo\")"

// Parse text
toolCalls := ParseToolCalls(response.Content)
```

### After (Native)
```go
// Prompt with tools
prompt.Tools = []ToolDefinition{{
    Type: "function",
    Function: FunctionDefinition{
        Name: "check_weather",
        Description: "Get weather",
        Parameters: {...},
    },
}}

// LLM response (structured)
response.ToolCalls = []ToolCallResponse{{
    Function: FunctionCallResponse{
        Name: "check_weather",
        Arguments: {"location": "Tokyo"},
    },
}}

// Direct execution (no parsing!)
result := tool.Execute(ctx, response.ToolCalls[0].Function.Arguments)
```

---

## Appendix B: Model Compatibility Matrix

| Model | Native Tools | Text-Based | Recommended |
|-------|-------------|------------|-------------|
| functiongemma:latest | ✅ | ⚠️ | Native |
| llama3.2 | ✅ | ✅ | Native |
| qwen2.5 | ✅ | ✅ | Native |
| mistral | ✅ | ✅ | Native |
| gemma3:latest | ❌ | ✅ | Text |
| gpt-4o | ✅ | ✅ | Native |
| claude-3 | ✅ | ✅ | Native |

---

**End of Analysis**
