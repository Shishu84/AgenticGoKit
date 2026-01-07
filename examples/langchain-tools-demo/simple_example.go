package main

import (
	"context"
	"fmt"
	"log"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"

	// Import plugins
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	_ "github.com/agenticgokit/agenticgokit/plugins/mcp/default"
)

// SIMPLE EXAMPLE: Agent automatically uses tools without custom handler!
// ========================================================================

func SimpleAgentWithNativeTools() {
	ctx := context.Background()

	// Create agent - tools are AUTOMATICALLY available!
	agent, err := vnext.NewBuilder("simple-agent").
		WithConfig(&vnext.Config{
			Name: "simple-agent",
			// The system prompt tells the LLM about tools
			SystemPrompt: "You are a helpful assistant. Use available tools when they can help answer questions.",
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.7,
			},
			Timeout: 30 * time.Second,
		}).
		// No custom handler needed! Tools work automatically
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// The agent can now use all registered native tools automatically!
	result, err := agent.Run(ctx, "What is 25 multiplied by 4?")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Agent response:", result.Content)
	fmt.Println("Tools used:", result.ToolsCalled)
}

func SimpleAgentWithMCPTools() {
	ctx := context.Background()

	// Define MCP server
	server := vnext.MCPServer{
		Name:    "web-tools",
		Type:    "http_sse",
		Address: "localhost",
		Port:    8080,
		Enabled: true,
	}

	// Create agent with MCP tools - automatically available!
	agent, err := vnext.NewBuilder("mcp-agent").
		WithConfig(&vnext.Config{
			Name:         "mcp-agent",
			SystemPrompt: "You are a helpful assistant. Use available tools when needed.",
			LLM: vnext.LLMConfig{
				Provider: "ollama",
				Model:    "gemma3:1b",
			},
		}).
		WithTools(vnext.WithMCP(server)). // Just add MCP server!
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Agent automatically uses MCP tools when relevant
	result, _ := agent.Run(ctx, "Search for latest Go news")
	fmt.Println("Response:", result.Content)
}

func SimpleAgentWithBothTools() {
	ctx := context.Background()

	// Agent with BOTH native and MCP tools - all automatic!
	agent, err := vnext.NewBuilder("full-agent").
		WithPreset(vnext.ChatAgent). // Use preset for simplicity
		WithTools(
			vnext.WithMCPDiscovery(8080, 8081), // Auto-discover MCP
		).
		Build()

	if err != nil {
		log.Fatal(err)
	}

	// Native tools (calculator, timestamp, text_processor) automatically available
	// MCP tools automatically discovered
	// Agent decides which tools to use based on the query!

	queries := []string{
		"Calculate 15 + 27",
		"What time is it?",
		"Reverse the text 'Hello World'",
	}

	for _, query := range queries {
		fmt.Printf("\nQuery: %s\n", query)
		result, _ := agent.Run(ctx, query)
		fmt.Printf("Response: %s\n", result.Content)
		if len(result.ToolsCalled) > 0 {
			fmt.Printf("Tools used: %v\n", result.ToolsCalled)
		}
	}
}

// This is even simpler - one line!
func SimplestPossibleExample() {
	ctx := context.Background()

	// One-liner agent with automatic tool support
	agent, _ := vnext.NewBuilder("agent").
		WithPreset(vnext.ChatAgent).
		Build()

	// All registered native tools are automatically available!
	result, _ := agent.Run(ctx, "Calculate 10 times 5")
	fmt.Println(result.Content)
}

func main() {
	fmt.Println("=== Simple Tool Usage Examples ===\n")

	fmt.Println("1. Agent with Native Tools (Automatic)")
	fmt.Println("---------------------------------------")
	SimpleAgentWithNativeTools()

	fmt.Println("\n2. Simplest Possible (One Line)")
	fmt.Println("---------------------------------------")
	SimplestPossibleExample()

	fmt.Println("\n3. Agent with Both Native + MCP Tools")
	fmt.Println("---------------------------------------")
	SimpleAgentWithBothTools()
}
