# Adding Tools to AgenticGoKit

**The Complete Guide to Tool Integration**

## 📚 Documentation Files

- **[GETTING_STARTED.md](GETTING_STARTED.md)** - Start here! Simplest approach ⭐
- **[COMPARISON.md](COMPARISON.md)** - Compare automatic vs manual approaches
- **[README_TOOLS.md](README_TOOLS.md)** - This file (complete reference)

## 🚀 Quick Start (Simplest Way)

### Native Tools work automatically - no custom handler needed!

```go
import (
    vnext "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
)

func main() {
    // That's it! All registered tools are automatically available
    agent, _ := vnext.NewBuilder("agent").
        WithPreset(vnext.ChatAgent).
        Build()
    
    // Agent automatically uses calculator, timestamp, text_processor tools
    result, _ := agent.Run(ctx, "Calculate 25 times 4")
    fmt.Println(result.Content)
    fmt.Println("Tools used:", result.ToolsCalled) // Shows which tools were called
}
```

### MCP Tools also work automatically!

```go
// Just add MCP servers - tools are auto-discovered and used
agent, _ := vnext.NewBuilder("agent").
    WithPreset(vnext.ChatAgent).
    WithTools(vnext.WithMCPDiscovery(8080, 8081)). // Auto-discover MCP servers
    Build()

// Agent automatically uses both native and MCP tools as needed!
result, _ := agent.Run(ctx, "Search for Go tutorials")
```

**That's it!** Tools are automatically available and the LLM decides when to use them. No custom handler needed!

---

## Available Tool Types

### 1. Native Go Tools (Recommended for Simple Operations)
Tools implemented directly in Go using the `v1beta.Tool` interface.

**Advantages:**
- ✅ No external dependencies
- ✅ Fast in-process execution
- ✅ Type-safe Go code
- ✅ Easy to test and debug

**Example Tools Included:**
- `calculator` - Basic arithmetic operations
- `get_timestamp` - Current timestamp in various formats
- `text_processor` - Text manipulation (uppercase, lowercase, reverse, count)

### 2. MCP (Model Context Protocol) Servers
External tool servers using the MCP protocol.

**Advantages:**
- ✅ Language-agnostic (Python, Node.js, etc.)
- ✅ Existing MCP ecosystem
- ✅ Complex external integrations
- ✅ Hot-reload capabilities

**Supported Transports:**
- `stdio` - Local process communication
- `http_sse` - HTTP Server-Sent Events
- `tcp` - TCP connections
- `websocket` - WebSocket connections

## How to Add Native Go Tools

### Method 1: Automatic (Recommended) ⭐

Tools are **automatically available** to all agents once registered. No extra code needed!

```go
// 1. Create and register your tool (in init or separate file)
type MyTool struct{}

func (t *MyTool) Name() string { return "my_tool" }
func (t *MyTool) Description() string { return "What it does" }
func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (*v1beta.ToolResult, error) {
    // Your logic here
    return &v1beta.ToolResult{Success: true, Content: result}, nil
}

func init() {
    v1beta.RegisterInternalTool("my_tool", func() v1beta.Tool { return &MyTool{} })
}

// 2. Create agent - tools automatically work!
agent, _ := v1beta.NewBuilder("agent").
    WithPreset(v1beta.ChatAgent).
    Build()

// 3. Use agent - it automatically calls tools when needed
result, _ := agent.Run(ctx, "Use my_tool with param xyz")
fmt.Println(result.ToolsCalled) // Shows: ["my_tool"]
```

### Method 2: Manual/Custom Handler (Advanced)

Use this only if you need custom logic for when/how to call tools.

### Step 1: Create Your Tool

Implement the `v1beta.Tool` interface:

```go
type MyCustomTool struct{}

func (t *MyCustomTool) Name() string {
    return "my_tool"
}

func (t *MyCustomTool) Description() string {
    return "Describe what your tool does"
}

func (t *MyCustomTool) Execute(ctx context.Context, args map[string]interface{}) (*v1beta.ToolResult, error) {
    // 1. Extract and validate arguments
    param, ok := args["param"].(string)
    if !ok {
        return &v1beta.ToolResult{
            Success: false,
            Error:   "missing required argument: param",
        }, fmt.Errorf("param required")
    }

    // 2. Perform your logic
    result := doSomething(param)

    // 3. Return result
    return &v1beta.ToolResult{
        Success: true,
        Content: map[string]interface{}{
            "output": result,
        },
    }, nil
}
```

### Step 2: Register Your Tool

Register in an `init()` function:

```go
func init() {
    v1beta.RegisterInternalTool("my_tool", func() v1beta.Tool {
        return &MyCustomTool{}
    })
}
```

### Step 3: Use Your Tool

**Option A: Automatic (Recommended)** ⭐

```go
// Just build the agent - tools work automatically!
agent, _ := v1beta.NewBuilder("agent").
    WithPreset(v1beta.ChatAgent).
    Build()

// Agent automatically uses registered tools based on user input
result, _ := agent.Run(ctx, "Use my_tool please")
fmt.Println("Tools called:", result.ToolsCalled)
```

**Option B: Direct Execution (Testing/Debugging)**

```go
// Discover all tools
tools, _ := v1beta.DiscoverTools()
for _, tool := range tools {
    fmt.Printf("- %s: %s\n", tool.Name(), tool.Description())
}

// Execute directly
result, _ := v1beta.ExecuteToolByName(ctx, "my_tool", map[string]interface{}{
    "param": "value",
})
```

**Option C: Custom Handler (Advanced)**

```go
// Use only if you need custom logic for tool selection
agent, _ := v1beta.NewBuilder("agent").
    WithHandler(func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
        if caps.Tools.IsAvailable("my_tool") {
            result, _ := caps.Tools.Execute(ctx, "my_tool", map[string]interface{}{
                "param": input,
            })
            // Use result...
        }
        return caps.LLM("system prompt", input)
    }).
    Build()
```

## How to Add MCP Servers

### Method 1: Explicit Server Configuration

```go
servers := []v1beta.MCPServer{
    {
        Name:    "filesystem",
        Type:    "stdio",
        Command: "npx",
        Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
        Enabled: true,
    },
    {
        Name:    "web-api",
        Type:    "http_sse",
        Address: "localhost",
        Port:    8080,
        Enabled: true,
    },
}

agent, _ := v1beta.NewBuilder("agent").
    WithTools(
        v1beta.WithMCP(servers...),
        v1beta.WithToolTimeout(30*time.Second),
    ).
    Build()
```

### Method 2: Auto-Discovery

```go
agent, _ := v1beta.NewBuilder("agent").
    WithTools(
        v1beta.WithMCPDiscovery(8080, 8081, 8090),
        v1beta.WithToolTimeout(30*time.Second),
    ).
    Build()
```

### Method 3: TOML Configuration

Create `config.toml`:

```toml
[tools]
enabled = true
timeout = "30s"

[[tools.mcp.servers]]
name = "filesystem"
type = "stdio"
command = "npx"
args = ["-y", "@modelcontextprotocol/server-filesystem"]
enabled = true

[[tools.mcp.servers]]
name = "web-search"
type = "http_sse"
address = "localhost"
port = 8080
enabled = true
```

Load in code:

```go
config, _ := v1beta.LoadConfigFromFile("config.toml")
agent, _ := v1beta.NewBuilder("agent").WithConfig(config).Build()
```

## Testing Your Tools

```go
func TestCalculatorTool(t *testing.T) {
    tool := &CalculatorTool{}
    
    ctx := context.Background()
    result, err := tool.Execute(ctx, map[string]interface{}{
        "operation": "add",
        "a":         10.0,
        "b":         5.0,
    })
    
    assert.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, 15.0, result.Content.(map[string]interface{})["result"])
}
```

## 📂 Example Files

### For Beginners (Start Here!)

1. **`getting_started.go`** - Simplest example, automatic tool usage ⭐
   ```bash
   go run getting_started.go custom_tools.go
   ```

2. **`simple_example.go`** - More automatic usage examples
   ```bash
   go run simple_example.go custom_tools.go
   ```

### For Reference

3. **`custom_tools.go`** - Example tool implementations (imported by other examples)

4. **`quick_reference.go`** - Code templates and patterns

### For Advanced Users

5. **`main_enhanced.go`** - Shows all three methods (automatic, direct, custom handler)
   ```bash
   go run main_enhanced.go custom_tools.go
   ```

6. **`main.go`** - Original MCP demo
   ```bash
   go run main.go custom_tools.go
   ```

## 🎯 Recommended Learning Path

1. Read **[GETTING_STARTED.md](GETTING_STARTED.md)** (5 min)
2. Run `getting_started.go` (1 min)
3. Read **[COMPARISON.md](COMPARISON.md)** (5 min)
4. Explore `simple_example.go` for more patterns
5. Check `main_enhanced.go` if you need custom handlers

## Next Steps

1. **Explore existing MCP servers**: https://github.com/modelcontextprotocol
2. **Create complex tools**: Add database access, API integrations, file operations
3. **Combine tools**: Use multiple tools together in workflows
4. **Tool caching**: Enable caching for expensive operations
5. **Monitoring**: Track tool usage and performance

## Common Patterns

### Conditional Tool Execution

```go
WithHandler(func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
    // Check if input needs calculation
    if strings.Contains(input, "calculate") {
        result, _ := caps.Tools.Execute(ctx, "calculator", ...)
        return formatResult(result), nil
    }
    
    // Fall back to LLM
    return caps.LLM("You are helpful.", input)
})
```

### Sequential Tool Calls

```go
// Get timestamp first
timeResult, _ := caps.Tools.Execute(ctx, "get_timestamp", map[string]interface{}{
    "format": "human",
})

// Then perform calculation
calcResult, _ := caps.Tools.Execute(ctx, "calculator", map[string]interface{}{
    "operation": "add",
    "a":         float64(timeResult.Content.(map[string]interface{})["timestamp"].(int64)),
    "b":         100.0,
})
```

### Error Handling

```go
result, err := caps.Tools.Execute(ctx, "calculator", args)
if err != nil {
    return fmt.Sprintf("Tool error: %v", err), nil
}

if !result.Success {
    return fmt.Sprintf("Tool failed: %s", result.Error), nil
}

// Use result.Content
```
