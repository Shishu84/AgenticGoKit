# Native Tool Calling: Quick Reference for LLM Drivers

A step-by-step checklist for adding native tool calling support to any LLM driver.

## Checklist

### 1. Core Types (`core/llm.go`)
- [x] `ToolDefinition` with Type and FunctionDefinition
- [x] `FunctionDefinition` with Name, Description, Parameters
- [x] `ToolCallResponse` with Type and FunctionCallResponse
- [x] `FunctionCallResponse` with Name and Arguments
- [x] Add `Tools []ToolDefinition` to `Prompt`
- [x] Add `ToolCalls []ToolCallResponse` to `Response`

### 2. LLM Adapter (`internal/llm/{provider}_adapter.go`)

#### Call Method
- [ ] Accept `prompt.Tools` as input
- [ ] Convert tools to provider-specific format
- [ ] Include tools in request payload
- [ ] Set `tool_choice: "auto"` or equivalent
- [ ] Parse `tool_calls` from response message
- [ ] Map provider format to core `ToolCallResponse`
- [ ] Return `Response.ToolCalls` populated

#### Stream Method (Optional)
- [ ] Include tools in streaming request
- [ ] Parse tool_call events from stream
- [ ] Emit tool calls as tokens or events
- [ ] Document limitations if streaming tools not supported

### 3. Plugin Wrapper (`plugins/llm/{provider}/{provider}.go`)
- [ ] Forward `core.Tools` to internal adapter
- [ ] Map internal `ToolCalls` back to public types
- [ ] Maintain backward compatibility

### 4. Agent Runtime (`v1beta/agent_impl.go`)
Already implemented; verify:
- [ ] `convertToolsToLLMFormat()` converts v1beta tools
- [ ] `getToolSchema()` extracts JSONSchema from tools
- [ ] `executeNativeToolsAndContinue()` runs structured calls
- [ ] Tool results fed back to LLM for reasoning

### 5. Configuration (`v1beta/config.go`, `v1beta/tools.go`)
Already implemented; available:
- [x] `ToolsConfig.SingleCallPolicy` ("best", "first", "all")
- [x] Default config in `DefaultToolsConfig()`

### 6. Testing
- [ ] Unit test: adapter with tools in request
- [ ] Unit test: parse tool_calls from response
- [ ] Integration test: end-to-end tool execution
- [ ] Edge case: empty tools, multiple calls, failures

### 7. Documentation
- [ ] Provider-specific tool format in README
- [ ] Example tool call and response
- [ ] Limitations (e.g., streaming, model support)

---

## Code Template

### Step 1: Update Adapter Call Method

```go
func (a *YourAdapter) Call(ctx context.Context, prompt llm.Prompt) (llm.Response, error) {
    // Build base request
    requestBody := map[string]interface{}{
        "model":    a.model,
        "messages": messages,
        "max_tokens": maxTokens,
        "temperature": temperature,
    }
    
    // ADD THIS: Include tools if present
    if len(prompt.Tools) > 0 {
        tools := make([]map[string]interface{}, len(prompt.Tools))
        for i, tool := range prompt.Tools {
            tools[i] = map[string]interface{}{
                "type": tool.Type,  // "function"
                "function": map[string]interface{}{
                    "name":        tool.Function.Name,
                    "description": tool.Function.Description,
                    "parameters":  tool.Function.Parameters,
                },
            }
        }
        requestBody["tools"] = tools
        requestBody["tool_choice"] = "auto"  // Let model decide when to use tools
    }
    
    // Make request (provider-specific)
    resp, err := a.makeRequest(ctx, requestBody)
    // ... handle error ...
    
    // Parse response
    var apiResp struct {
        Message struct {
            Content   string                    `json:"content"`
            ToolCalls []map[string]interface{} `json:"tool_calls,omitempty"`
        } `json:"message"`
    }
    json.NewDecoder(resp.Body).Decode(&apiResp)
    
    // Build core response
    response := llm.Response{
        Content: apiResp.Message.Content,
    }
    
    // ADD THIS: Map tool calls to core types
    if len(apiResp.Message.ToolCalls) > 0 {
        response.ToolCalls = make([]llm.ToolCallResponse, len(apiResp.Message.ToolCalls))
        for i, tc := range apiResp.Message.ToolCalls {
            // Extract function name and arguments from provider format
            fn := tc["function"].(map[string]interface{})
            args := fn["arguments"].(map[string]interface{})
            
            response.ToolCalls[i] = llm.ToolCallResponse{
                Type: "function",
                Function: llm.FunctionCallResponse{
                    Name:      fn["name"].(string),
                    Arguments: args,
                },
            }
        }
    }
    
    return response, nil
}
```

### Step 2: Test Tool Calling

```go
func TestYourAdapterTools(t *testing.T) {
    adapter := NewYourAdapter(apiKey, model, maxTokens, temperature)
    
    prompt := llm.Prompt{
        System: "You are helpful.",
        User:   "Should you use the weather tool?",
        Tools: []llm.ToolDefinition{
            {
                Type: "function",
                Function: llm.FunctionDefinition{
                    Name:        "get_weather",
                    Description: "Get weather for a location",
                    Parameters: map[string]interface{}{
                        "type": "object",
                        "properties": map[string]interface{}{
                            "location": map[string]interface{}{
                                "type":        "string",
                                "description": "City name",
                            },
                        },
                        "required": []string{"location"},
                    },
                },
            },
        },
    }
    
    resp, err := adapter.Call(context.Background(), prompt)
    
    if err != nil {
        t.Fatalf("Call failed: %v", err)
    }
    
    // Verify tool calls are parsed
    if len(resp.ToolCalls) == 0 {
        t.Log("No tool calls in response (model may not have chosen to use tools)")
    } else {
        tc := resp.ToolCalls[0]
        if tc.Function.Name != "get_weather" {
            t.Errorf("Expected get_weather, got %s", tc.Function.Name)
        }
        if _, ok := tc.Function.Arguments["location"]; !ok {
            t.Error("Missing location argument")
        }
    }
}
```

---

## Provider-Specific Mappings

### Provider Request Format

| Provider | Tools Key | Tool Choice | Function Key | Args Key |
|----------|-----------|-------------|--------------|----------|
| OpenAI | `tools` | `tool_choice` | `function` | `function.arguments` |
| Anthropic | `tools` | N/A (auto) | `function` | `input` |
| Ollama | `tools` | `tool_choice` | `function` | `arguments` |
| Claude | `tools` | N/A | `name` | `input` |

### Provider Response Format

| Provider | Tool Calls Path | Tool Name | Arguments |
|----------|-----------------|-----------|-----------|
| OpenAI | `message.tool_calls[].function` | `.name` | `.arguments` (string) |
| Anthropic | `content[].type == "tool_use"` | `.name` | `.input` (object) |
| Ollama | `message.tool_calls[].function` | `.name` | `.arguments` (object) |
| Claude | Content blocks | `.name` | `.input` (object) |

---

## Common Pitfalls

### 1. Arguments Type Mismatch
Some providers return arguments as JSON string, others as objects. Always normalize:

```go
var args map[string]interface{}
if argsStr, ok := tc.Function.Arguments.(string); ok {
    json.Unmarshal([]byte(argsStr), &args)  // Parse string
} else {
    args = tc.Function.Arguments.(map[string]interface{})  // Already object
}
```

### 2. Forgetting tool_choice Field
Without `tool_choice: "auto"`, the model may ignore tools. Include it explicitly.

### 3. Not Normalizing Schema Format
Parameters should always be valid JSON Schema. Validate with provider docs.

### 4. Streaming Without Tool Support
Current streaming implementations may not emit tool_call events. Document this limitation and fall back to post-stream parsing.

---

## Verification Checklist

- [ ] Adapter accepts `prompt.Tools`
- [ ] Request includes tools array with correct format
- [ ] Model response parsed for `tool_calls`
- [ ] Core `ToolCallResponse` types populated
- [ ] Test passes with mocked/real model
- [ ] Tool calls propagate to agent runtime
- [ ] Agent executes tools and processes results
- [ ] Example or documentation updated
- [ ] No breaking changes to existing code

---

## Example Output

After implementing, you should see:

```bash
$ go test internal/llm/your_adapter_test.go -v
--- PASS: TestYourAdapterTools (5.23s)
    your_adapter_test.go:45: Tool calls parsed successfully
PASS

$ go run examples/langchain-weather-demo/main.go
   Assistant: It's always sunny in sf ☀️
```

---

## References

- Main Implementation: [docs/NATIVE_TOOL_CALLING_IMPLEMENTATION.md](NATIVE_TOOL_CALLING_IMPLEMENTATION.md)
- Core Types: [core/llm.go](../core/llm.go)
- Ollama Example: [internal/llm/ollama_adapter.go](../internal/llm/ollama_adapter.go)
- OpenAI Example: [internal/llm/openai_adapter.go](../internal/llm/openai_adapter.go)
- Agent Runtime: [v1beta/agent_impl.go](../v1beta/agent_impl.go#L1300)
- Example Demo: [examples/langchain-weather-demo/main.go](../examples/langchain-weather-demo/main.go)

---

**Questions?** See the full implementation guide above or check existing adapter implementations.
