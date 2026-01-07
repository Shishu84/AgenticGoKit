# Tool Integration Examples for AgenticGoKit

**Learn how to add tools to AgenticGoKit agents - the simple way!**

## 🚀 Start Here

👉 **New to tools?** Read [GETTING_STARTED.md](GETTING_STARTED.md) (5 min tutorial) ⭐

👉 **Need quick reference?** See [CHEATSHEET.md](CHEATSHEET.md)

👉 **Want to compare approaches?** See [COMPARISON.md](COMPARISON.md)

## ⚡ 30-Second Quick Start

```go
// 1. Create and register your tool
type MyTool struct{}
func (t *MyTool) Name() string { return "my_tool" }
func (t *MyTool) Description() string { return "What it does" }
func (t *MyTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
    return &vnext.ToolResult{Success: true, Content: result}, nil
}
func init() {
    vnext.RegisterInternalTool("my_tool", func() vnext.Tool { return &MyTool{} })
}

// 2. Use it automatically - that's it!
agent, _ := vnext.NewBuilder("agent").WithPreset(vnext.ChatAgent).Build()
result, _ := agent.Run(ctx, "Use my_tool")
```

## 📂 Example Files

| File | Description | Run |
|------|-------------|-----|
| **`getting_started.go`** | Simplest automatic usage ⭐ | `go run getting_started.go custom_tools.go` |
| **`simple_example.go`** | More automatic patterns | `go run simple_example.go custom_tools.go` |
| **`main_enhanced.go`** | All three methods | `go run main_enhanced.go custom_tools.go` |
| **`main.go`** | Original MCP demo | `go run main.go custom_tools.go` |

## 🎯 What This Demo Shows

- ✅ **Native Go Tools**: Create custom tools in pure Go
- ✅ **Automatic Tool Usage**: Tools work without custom handlers
- ✅ **MCP Integration**: Connect to external MCP servers
- ✅ **Three Usage Patterns**: Automatic, Direct, and Custom Handler

## Prerequisites

- Go 1.21+
- An LLM provider; this demo uses Ollama with `gemma3:1b`
  - `ollama pull gemma3:1b`
- Optional: MCP server for external tools (auto-discovered on ports 8080, 8081, etc.)

## 🧰 Included Tools

Three ready-to-use tools in `custom_tools.go`:

1. **calculator** - Math operations (add, subtract, multiply, divide)
2. **get_timestamp** - Current time in various formats
3. **text_processor** - Text operations (uppercase, lowercase, reverse, count)

## 📚 Documentation

- **[GETTING_STARTED.md](GETTING_STARTED.md)** - Beginner tutorial ⭐
- **[CHEATSHEET.md](CHEATSHEET.md)** - Quick reference & templates
- **[COMPARISON.md](COMPARISON.md)** - Compare automatic vs manual approaches
- **[README_TOOLS.md](README_TOOLS.md)** - Complete documentation

## 🎓 Learning Path

1. Read [GETTING_STARTED.md](GETTING_STARTED.md) (5 min)
2. Run `go run getting_started.go custom_tools.go` (1 min)
3. Explore `simple_example.go` for more patterns
4. Check [CHEATSHEET.md](CHEATSHEET.md) for templates

## Run it (Windows PowerShell)

```powershell
# From repo root
pwsh -NoProfile -Command "cd examples/vnext/mcp-tools-blog-demo; go run ."
```

## Notes

- Plugins are imported blank inside `main.go`:
  - `plugins/mcp/unified` (transport)
  - `plugins/mcp/default` (registry/cache)
  - `plugins/llm/ollama` (LLM provider)
- Swap the LLM provider/model to match your environment if you prefer OpenAI/Azure/etc.
