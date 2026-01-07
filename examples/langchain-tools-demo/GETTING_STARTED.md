# Getting Started - Automatic Tool Usage

**The simplest way to use tools in AgenticGoKit!**

## ⚡ Quick Start (3 Steps)

### Step 1: Create your tool (custom_tools.go)

```go
package main

import (
    "context"
    vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

type CalculatorTool struct{}

func (t *CalculatorTool) Name() string { 
    return "calculator" 
}

func (t *CalculatorTool) Description() string { 
    return "Performs math operations: add, subtract, multiply, divide" 
}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
    // Extract args, do calculation, return result
    operation := args["operation"].(string)
    a := args["a"].(float64)
    b := args["b"].(float64)
    
    var result float64
    switch operation {
    case "multiply":
        result = a * b
    // ... other operations
    }
    
    return &vnext.ToolResult{
        Success: true,
        Content: map[string]interface{}{"result": result},
    }, nil
}

// Register the tool
func init() {
    vnext.RegisterInternalTool("calculator", func() vnext.Tool {
        return &CalculatorTool{}
    })
}
```

### Step 2: Create agent (main.go)

```go
package main

import (
    "context"
    vnext "github.com/agenticgokit/agenticgokit/v1beta"
    _ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
)

func main() {
    ctx := context.Background()
    
    // Just use a preset - tools work automatically!
    agent, _ := vnext.NewBuilder("agent").
        WithPreset(vnext.ChatAgent).
        Build()
    
    // Ask anything - agent uses tools automatically
    result, _ := agent.Run(ctx, "Calculate 25 times 4")
    
    fmt.Println("Response:", result.Content)
    fmt.Println("Tools used:", result.ToolsCalled) // ["calculator"]
}
```

### Step 3: Run

```bash
go run main.go custom_tools.go
```

**That's it!** The agent automatically:
- Detects when tools are needed
- Chooses the right tool
- Calls it with correct arguments
- Returns formatted results

## 🎯 Key Points

### ✅ Tools Work Automatically
- No custom handler code needed
- No manual tool checking
- LLM decides when to use tools
- Works with any preset (ChatAgent, ResearchAgent, etc.)

### ✅ Simple Registration
```go
func init() {
    vnext.RegisterInternalTool("tool_name", func() vnext.Tool {
        return &YourTool{}
    })
}
```

### ✅ See What Tools Were Called
```go
result, _ := agent.Run(ctx, "your question")
fmt.Println(result.ToolsCalled) // Shows which tools were used
```

## 📋 Complete Example

See `getting_started.go` for a complete working example with multiple tools.

```bash
go run getting_started.go custom_tools.go
```

## 🔧 Adding MCP Tools (Also Automatic!)

```go
// MCP tools are also automatic
agent, _ := vnext.NewBuilder("agent").
    WithPreset(vnext.ChatAgent).
    WithTools(
        vnext.WithMCPDiscovery(8080, 8081), // Auto-discover
    ).
    Build()

// Works the same - agent uses MCP tools automatically
result, _ := agent.Run(ctx, "Search for news")
```

## 🆚 When to Use Custom Handlers?

**Use Automatic (Recommended for 90% of cases)**
- Simple tool usage
- Let LLM decide when to use tools
- Clean, minimal code

**Use Custom Handler (Advanced)**
- Complex tool orchestration
- Multi-step tool workflows
- Conditional tool logic
- Tool result transformation

## 💡 Tips

1. **Good tool descriptions help**: LLM uses descriptions to decide when to call tools
2. **Clear argument names**: Use descriptive parameter names
3. **Check result.ToolsCalled**: Debug which tools were used
4. **Start simple**: Begin with automatic usage, add custom handlers only if needed

## 📚 Next Steps

- See `simple_example.go` for more automatic usage examples
- See `main_enhanced.go` for custom handler examples
- See `custom_tools.go` for more tool implementations
- Read `README_TOOLS.md` for complete documentation
