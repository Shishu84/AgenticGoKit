# Tool Integration Methods - Comparison

## 🎯 Three Ways to Use Tools

### 1️⃣ Automatic (Recommended) ⭐

**When**: 90% of use cases  
**Complexity**: Simplest  
**Code**: Minimal

```go
// Just build and run - tools work automatically!
agent, _ := vnext.NewBuilder("agent").
    WithPreset(vnext.ChatAgent).
    Build()

result, _ := agent.Run(ctx, "Calculate 10 + 5")
// Agent automatically calls calculator tool
```

**Pros:**
- ✅ No custom code
- ✅ LLM chooses when to use tools
- ✅ Clean and maintainable
- ✅ Works with all tools (native + MCP)

**Cons:**
- ❌ Less control over tool selection
- ❌ Can't customize tool orchestration

---

### 2️⃣ Direct Execution (For Testing)

**When**: Testing, debugging, scripting  
**Complexity**: Simple  
**Code**: Minimal

```go
// Execute tools directly (bypass LLM)
result, _ := vnext.ExecuteToolByName(ctx, "calculator", map[string]interface{}{
    "operation": "add",
    "a": 10.0,
    "b": 5.0,
})
```

**Pros:**
- ✅ Direct tool access
- ✅ No LLM needed
- ✅ Fast and deterministic
- ✅ Good for testing

**Cons:**
- ❌ Manual tool selection
- ❌ Manual argument preparation

---

### 3️⃣ Custom Handler (Advanced)

**When**: Complex workflows, custom logic  
**Complexity**: Most complex  
**Code**: More code required

```go
agent, _ := vnext.NewBuilder("agent").
    WithHandler(func(ctx context.Context, input string, caps *v1beta.Capabilities) (string, error) {
        // Custom logic for tool selection
        if strings.Contains(input, "calculate") {
            result, _ := caps.Tools.Execute(ctx, "calculator", args)
            return formatResult(result), nil
        }
        return caps.LLM("prompt", input)
    }).
    Build()
```

**Pros:**
- ✅ Full control over tool usage
- ✅ Custom orchestration logic
- ✅ Multi-step workflows
- ✅ Result transformation

**Cons:**
- ❌ More code to maintain
- ❌ Need to handle tool selection manually
- ❌ More complex to debug

---

## 📊 Quick Comparison

| Feature | Automatic | Direct Execution | Custom Handler |
|---------|-----------|------------------|----------------|
| **Simplicity** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ |
| **Control** | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Code Required** | Minimal | Minimal | Most |
| **LLM Integration** | Yes | No | Yes |
| **Good for Beginners** | ✅ Yes | ✅ Yes | ❌ No |
| **Production Ready** | ✅ Yes | ⚠️ Limited | ✅ Yes |

---

## 🎓 Which Should I Use?

### Start with Automatic ⭐

```go
// This covers 90% of use cases!
agent, _ := vnext.NewBuilder("agent").
    WithPreset(vnext.ChatAgent).
    Build()

result, _ := agent.Run(ctx, "your question")
```

### Use Direct Execution for:
- ✅ Unit tests
- ✅ Tool debugging
- ✅ Scripts/automation
- ✅ Non-AI workflows

### Use Custom Handler for:
- ✅ Multi-step tool workflows
- ✅ Conditional tool logic
- ✅ Result transformations
- ✅ Complex orchestrations

---

## 💡 Pro Tips

1. **Start Simple**: Use automatic mode first
2. **Add Handlers Later**: Only when you need custom logic
3. **Mix Approaches**: Use direct execution for testing, automatic for production
4. **Check ToolsCalled**: See what the agent did automatically

```go
result, _ := agent.Run(ctx, "question")
fmt.Println(result.ToolsCalled) // ["calculator", "timestamp"]
```

---

## 📝 Examples

### All Three Methods in One Program

```go
package main

func main() {
    ctx := context.Background()
    
    // Method 1: Automatic
    agent1, _ := vnext.NewBuilder("auto").WithPreset(vnext.ChatAgent).Build()
    r1, _ := agent1.Run(ctx, "Calculate 5 + 3")
    
    // Method 2: Direct
    r2, _ := vnext.ExecuteToolByName(ctx, "calculator", map[string]interface{}{
        "operation": "add", "a": 5.0, "b": 3.0,
    })
    
    // Method 3: Custom
    agent3, _ := vnext.NewBuilder("custom").
        WithHandler(func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
            result, _ := caps.Tools.Execute(ctx, "calculator", args)
            return fmt.Sprintf("Result: %v", result.Content), nil
        }).
        Build()
    r3, _ := agent3.Run(ctx, "anything")
}
```

---

## 🚀 Recommendation

**For New Users:**
```go
// Just do this! 
agent, _ := vnext.NewBuilder("agent").WithPreset(vnext.ChatAgent).Build()
result, _ := agent.Run(ctx, "your question")
```

**For Advanced Users:**
- Read `main_enhanced.go` for custom handler examples
- Read `simple_example.go` for automatic usage patterns
- Read `quick_reference.go` for code templates

---

**Remember**: Simpler is better. Use automatic mode unless you have a specific reason not to! 🎯
