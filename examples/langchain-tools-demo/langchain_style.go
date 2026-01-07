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
// LANGCHAIN-STYLE SIMPLE TOOL DEFINITION
// ============================================================================

// WeatherTool - Simple function-based tool (LangChain style)
type WeatherTool struct{}

func (t *WeatherTool) Name() string {
	return "check_weather"
}

func (t *WeatherTool) Description() string {
	return "Return the weather forecast for the specified location"
}

func (t *WeatherTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	location, ok := args["location"].(string)
	if !ok {
		return &vnext.ToolResult{
			Success: false,
			Error:   "location parameter required",
		}, fmt.Errorf("location required")
	}

	// Simple weather response (you can replace with actual API call)
	weather := fmt.Sprintf("It's always sunny in %s ☀️", location)

	return &vnext.ToolResult{
		Success: true,
		Content: map[string]interface{}{
			"location": location,
			"forecast": weather,
		},
	}, nil
}

// Register the tool (automatically available to all agents)
func init() {
	vnext.RegisterInternalTool("check_weather", func() vnext.Tool {
		return &WeatherTool{}
	})
}

// ============================================================================
// LANGCHAIN-STYLE SIMPLE AGENT CREATION
// ============================================================================

func main() {
	ctx := context.Background()

	// LangChain-style: Simple agent creation with automatic tool support
	// graph = create_agent(model="...", tools=[check_weather], system_prompt="...")
	agent, err := vnext.NewBuilder("weather-agent").
		WithConfig(&vnext.Config{
			Name:         "weather-agent",
			SystemPrompt: "You are a helpful assistant",
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.7,
				MaxTokens:   150,
			},
			Timeout: 30 * time.Second,
		}).
		WithPreset(vnext.ChatAgent). // Tools automatically available
		Build()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("🌤️  Weather Agent - LangChain Style Example")
	fmt.Println("==========================================\n")

	// LangChain-style: inputs = {"messages": [{"role": "user", "content": "..."}]}
	queries := []string{
		"what is the weather in sf",
		"how's the weather in New York?",
		"tell me the forecast for Tokyo",
		"what's it like in London today?",
	}

	for i, query := range queries {
		fmt.Printf("%d. User: %s\n", i+1, query)

		// Simple run (non-streaming)
		result, err := agent.Run(ctx, query)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("   🤖 Assistant: %s\n", result.Content)
		if len(result.ToolsCalled) > 0 {
			fmt.Printf("   🔧 Tools used: %v\n", result.ToolsCalled)
		}
		fmt.Printf("   ⏱️  Duration: %v\n\n", result.Duration)
	}

	fmt.Println("\n" + "=".repeat(60))
	fmt.Println("🌟 Streaming Example (LangChain-style)")
	fmt.Println("=".repeat(60) + "\n")

	// LangChain-style streaming: for chunk in graph.stream(inputs, stream_mode="updates")
	streamingExample(ctx, agent)
}

// Streaming example similar to LangChain's graph.stream()
func streamingExample(ctx context.Context, agent vnext.Agent) {
	fmt.Println("User: what is the weather in San Francisco?\n")
	fmt.Print("Assistant: ")

	// Stream the response
	stream, err := agent.RunStream(ctx, "what is the weather in San Francisco?")
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
	defer stream.Close()

	// Process chunks (similar to LangChain's streaming)
	for {
		chunk, err := stream.Recv()
		if err != nil {
			break
		}

		// Print content as it arrives
		if chunk.Type == "content" {
			fmt.Print(chunk.Content)
		}

		// Show tool calls
		if chunk.Type == "tool_call" {
			fmt.Printf("\n   [Calling tool: %s]\n   ", chunk.ToolName)
		}

		// Show tool results
		if chunk.Type == "tool_result" {
			fmt.Printf("\n   [Tool result: %v]\n   ", chunk.Content)
		}
	}

	fmt.Println("\n")
}
