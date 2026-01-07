# Tool Integration Cheat Sheet

## ⚡ Super Quick Start

### 1. Create Tool (3 methods)
```go
type MyTool struct{}
func (t *MyTool) Name() string { return "my_tool" }
func (t *MyTool) Description() string { return "What it does" }
func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
    return &vnext.ToolResult{Success: true, Content: result}, nil
}
```

### 2. Register Tool
```go
func init() {
    vnext.RegisterInternalTool("my_tool", func() vnext.Tool { return &MyTool{} })
}
```

### 3. Use Tool (Choose ONE)

#### Option A: Automatic ⭐ (Recommended)
```go
agent, _ := vnext.NewBuilder("agent").WithPreset(vnext.ChatAgent).Build()
result, _ := agent.Run(ctx, "Use my_tool")
// Tools are called automatically!
```

#### Option B: Direct
```go
result, _ := vnext.ExecuteToolByName(ctx, "my_tool", args)
```

#### Option C: Custom Handler
```go
agent, _ := vnext.NewBuilder("agent").
    WithHandler(func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
        result, _ := caps.Tools.Execute(ctx, "my_tool", args)
        return format(result), nil
    }).Build()
```

---

## 🔧 Common Patterns

### Native Tool
```go
func init() {
    vnext.RegisterInternalTool("calculator", func() vnext.Tool {
        return &CalculatorTool{}
    })
}
```

### MCP Server (Automatic)
```go
agent, _ := vnext.NewBuilder("agent").
    WithTools(vnext.WithMCPDiscovery(8080, 8081)).
    Build()
```

### MCP Server (Explicit)
```go
server := vnext.MCPServer{Name: "fs", Type: "stdio", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem"}}
agent, _ := vnext.NewBuilder("agent").WithTools(vnext.WithMCP(server)).Build()
```

### Check Which Tools Were Used
```go
result, _ := agent.Run(ctx, "question")
fmt.Println(result.ToolsCalled) // ["tool1", "tool2"]
```

### Discover Available Tools
```go
tools, _ := vnext.DiscoverTools()
for _, t := range tools {
    fmt.Printf("%s: %s\n", t.Name(), t.Description())
}
```

---

## 📋 Tool Template

```go
package main

import (
    "context"
    vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// 1. Define struct
type MyTool struct {
    // Optional: config fields
}

// 2. Implement Name
func (t *MyTool) Name() string {
    return "my_tool_name"
}

// 3. Implement Description
func (t *MyTool) Description() string {
    return "Clear description. Args: param1 (type), param2 (type)"
}

// 4. Implement Execute
func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
    // Extract args
    param, ok := args["param"].(string)
    if !ok {
        return &vnext.ToolResult{Success: false, Error: "param required"}, nil
    }
    
    // Do work
    result := doWork(param)
    
    // Return
    return &vnext.ToolResult{
        Success: true,
        Content: map[string]interface{}{"output": result},
    }, nil
}

// 5. Register in init
func init() {
    vnext.RegisterInternalTool("my_tool_name", func() vnext.Tool {
        return &MyTool{}
    })
}
```

---

## 🎯 Decision Tree

```
Do you need tools?
├─ Yes
│  ├─ Simple usage? → Use Automatic ⭐
│  ├─ Testing only? → Use Direct Execution
│  └─ Complex logic? → Use Custom Handler
└─ No
   └─ Just use agent.Run()
```

---

## ⚠️ Common Mistakes

❌ **Don't**: Write custom handler for simple tool usage  
✅ **Do**: Use automatic mode (WithPreset)

❌ **Don't**: Manually check and call tools  
✅ **Do**: Let the LLM decide when to use tools

❌ **Don't**: Forget to call RegisterInternalTool  
✅ **Do**: Register in init()

❌ **Don't**: Use vague tool descriptions  
✅ **Do**: Write clear descriptions with argument types

---

## 🚀 Examples

```bash
# Automatic usage (simplest)
go run getting_started.go custom_tools.go

# More examples
go run simple_example.go custom_tools.go

# All three methods
go run main_enhanced.go custom_tools.go
```

---

## 📚 Learn More

- **[GETTING_STARTED.md](GETTING_STARTED.md)** - Tutorial
- **[COMPARISON.md](COMPARISON.md)** - Compare approaches
- **[README_TOOLS.md](README_TOOLS.md)** - Full reference

---

**Remember**: Start with automatic mode! 🎯

```go
agent, _ := vnext.NewBuilder("agent").WithPreset(vnext.ChatAgent).Build()
result, _ := agent.Run(ctx, "your question")
```
