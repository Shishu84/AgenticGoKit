package main

import (
	"context"
	"fmt"
	"log"
	"time"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// ============================================================================
// HELPER: Function to Tool Wrapper (makes it LangChain-style simple!)
// ============================================================================

// FuncTool wraps a simple function as a Tool
type FuncTool struct {
	name        string
	description string
	fn          func(args map[string]interface{}) (interface{}, error)
}

func (t *FuncTool) Name() string        { return t.name }
func (t *FuncTool) Description() string { return t.description }
func (t *FuncTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	result, err := t.fn(args)
	if err != nil {
		return &vnext.ToolResult{Success: false, Error: err.Error()}, err
	}
	return &vnext.ToolResult{Success: true, Content: result}, nil
}

// CreateTool - LangChain-style tool creation from a simple function
func CreateTool(name, description string, fn func(args map[string]interface{}) (interface{}, error)) vnext.Tool {
	return &FuncTool{
		name:        name,
		description: description,
		fn:          fn,
	}
}

// RegisterFuncTool - Register a function as a tool (LangChain-style)
func RegisterFuncTool(name, description string, fn func(args map[string]interface{}) (interface{}, error)) {
	vnext.RegisterInternalTool(name, func() vnext.Tool {
		return CreateTool(name, description, fn)
	})
}

// ============================================================================
// LANGCHAIN-STYLE EXAMPLE: Define tool as a simple function!
// ============================================================================

// Check weather - just a simple function!
func checkWeather(args map[string]interface{}) (interface{}, error) {
	location, ok := args["location"].(string)
	if !ok {
		return nil, fmt.Errorf("location parameter required")
	}
	return fmt.Sprintf("It's always sunny in %s ☀️", location), nil
}

// Get current time - another simple function!
func getCurrentTime(args map[string]interface{}) (interface{}, error) {
	format := "3:04 PM"
	if f, ok := args["format"].(string); ok {
		format = f
	}
	return time.Now().Format(format), nil
}

// Register tools (like LangChain's tools=[...])
func init() {
	RegisterFuncTool("check_weather", "Return the weather forecast for the specified location", checkWeather)
	RegisterFuncTool("get_time", "Get the current time", getCurrentTime)
}

// ============================================================================
// LANGCHAIN-STYLE AGENT USAGE
// ============================================================================

func main() {
	ctx := context.Background()

	fmt.Println("🚀 LangChain-Style Tool Usage in Go")
	fmt.Println("===================================\n")

	// Step 1: Create agent (LangChain: create_agent)
	agent, err := createAgent(
		"ollama:gemma3:1b",
		"You are a helpful assistant",
	)
	if err != nil {
		log.Fatal(err)
	}

	// Step 2: Run queries (LangChain: graph.stream)
	queries := []string{
		"what is the weather in sf",
		"what time is it?",
		"how's the weather in Tokyo?",
	}

	for i, query := range queries {
		fmt.Printf("%d. 👤 User: %s\n", i+1, query)

		result, err := agent.Run(ctx, query)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("   🤖 Assistant: %s\n", result.Content)
		if len(result.ToolsCalled) > 0 {
			fmt.Printf("   🔧 Tools: %v\n", result.ToolsCalled)
		}
		fmt.Println()
	}

	// Step 3: Streaming example (LangChain: graph.stream with chunks)
	fmt.Println("\n" + "─".repeat(60))
	fmt.Println("📡 Streaming Example")
	fmt.Println("─".repeat(60) + "\n")

	streamExample(ctx, agent, "what is the weather in San Francisco?")
}

// createAgent - LangChain-style agent creation
// Similar to: create_agent(model="...", tools=[...], system_prompt="...")
func createAgent(model, systemPrompt string) (vnext.Agent, error) {
	// Parse model string (format: "provider:model")
	provider := "ollama"
	modelName := "gemma3:1b"
	// You could parse model string here if needed

	return vnext.NewBuilder("agent").
		WithConfig(&vnext.Config{
			Name:         "agent",
			SystemPrompt: systemPrompt,
			LLM: vnext.LLMConfig{
				Provider:    provider,
				Model:       modelName,
				Temperature: 0.7,
				MaxTokens:   200,
			},
			Timeout: 30 * time.Second,
		}).
		WithPreset(vnext.ChatAgent). // Tools automatically available!
		Build()
}

// streamExample - LangChain-style streaming
// Similar to: for chunk in graph.stream(inputs, stream_mode="updates")
func streamExample(ctx context.Context, agent vnext.Agent, query string) {
	fmt.Printf("👤 User: %s\n\n", query)
	fmt.Print("🤖 Assistant: ")

	stream, err := agent.RunStream(ctx, query)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	defer stream.Close()

	toolsCalled := []string{}

	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}

		switch chunk.Type {
		case "content":
			fmt.Print(chunk.Content)

		case "tool_call":
			toolsCalled = append(toolsCalled, chunk.ToolName)
			fmt.Printf("\n   [🔧 Calling: %s]\n   ", chunk.ToolName)

		case "tool_result":
			fmt.Printf("\n   [✅ Result: %v]\n   ", chunk.Content)
		}
	}

	fmt.Println()
	if len(toolsCalled) > 0 {
		fmt.Printf("\n🔧 Tools used: %v\n", toolsCalled)
	}
	fmt.Println()
}

// ============================================================================
// EVEN SIMPLER: One-liner tool definition!
// ============================================================================

func exampleSimplestPossible() {
	// Define and register in one line!
	RegisterFuncTool("greet", "Greet a person", func(args map[string]interface{}) (interface{}, error) {
		name := args["name"].(string)
		return fmt.Sprintf("Hello, %s! 👋", name), nil
	})

	// Use it
	agent, _ := createAgent("ollama:gemma3:1b", "You are friendly")
	result, _ := agent.Run(context.Background(), "greet Alice")
	fmt.Println(result.Content)
}
