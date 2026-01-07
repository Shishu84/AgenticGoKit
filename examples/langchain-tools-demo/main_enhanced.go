package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"

	// MCP plugins: provide the manager/transport + registry
	_ "github.com/agenticgokit/agenticgokit/plugins/mcp/default"
	_ "github.com/agenticgokit/agenticgokit/plugins/mcp/unified"

	// LLM provider plugin
	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
)

func main() {
	fmt.Println("=== AgenticGoKit - Native + MCP Tools Demo ===\n")

	// Demo 1: Show all available tools (native + MCP)
	if err := demoToolDiscovery(); err != nil {
		log.Printf("Tool discovery demo error: %v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("-", 60) + "\n")

	// Demo 2: Use native tools directly
	if err := demoNativeTools(); err != nil {
		log.Printf("Native tools demo error: %v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("-", 60) + "\n")

	// Demo 3: Use MCP server tools
	if err := demoMCPTools(); err != nil {
		log.Printf("MCP tools demo error: %v\n", err)
	}

	fmt.Println("\n" + strings.Repeat("-", 60) + "\n")

	// Demo 4: Agent with both native and MCP tools
	if err := demoAgentWithTools(); err != nil {
		log.Printf("Agent demo error: %v\n", err)
	}
}

// demoToolDiscovery shows all available tools
func demoToolDiscovery() error {
	fmt.Println("📋 Tool Discovery Demo")
	fmt.Println("Discovering all available tools (native + MCP)...\n")

	ctx := context.Background()

	// Discover all tools
	tools, err := vnext.DiscoverTools()
	if err != nil {
		return fmt.Errorf("failed to discover tools: %w", err)
	}

	fmt.Printf("✅ Found %d tools:\n\n", len(tools))
	for i, tool := range tools {
		fmt.Printf("%d. %s\n", i+1, tool.Name())
		fmt.Printf("   📝 %s\n\n", tool.Description())
	}

	return nil
}

// demoNativeTools demonstrates using native Go tools
func demoNativeTools() error {
	fmt.Println("🔧 Native Tools Demo")
	fmt.Println("Testing custom Go tools...\n")

	ctx := context.Background()

	// Test 1: Calculator
	fmt.Println("1️⃣  Calculator Tool:")
	calcResult, err := vnext.ExecuteToolByName(ctx, "calculator", map[string]interface{}{
		"operation": "multiply",
		"a":         12.0,
		"b":         8.0,
	})
	if err != nil {
		return fmt.Errorf("calculator error: %w", err)
	}
	fmt.Printf("   Input: 12 × 8\n")
	fmt.Printf("   Result: %v\n", calcResult.Content)
	fmt.Printf("   Success: %v\n\n", calcResult.Success)

	// Test 2: Timestamp
	fmt.Println("2️⃣  Timestamp Tool:")
	timeResult, err := vnext.ExecuteToolByName(ctx, "get_timestamp", map[string]interface{}{
		"format": "human",
	})
	if err != nil {
		return fmt.Errorf("timestamp error: %w", err)
	}
	fmt.Printf("   Format: human-readable\n")
	fmt.Printf("   Result: %v\n", timeResult.Content)
	fmt.Printf("   Success: %v\n\n", timeResult.Success)

	// Test 3: Text Processor
	fmt.Println("3️⃣  Text Processor Tool:")
	textResult, err := vnext.ExecuteToolByName(ctx, "text_processor", map[string]interface{}{
		"operation": "reverse",
		"text":      "AgenticGoKit",
	})
	if err != nil {
		return fmt.Errorf("text processor error: %w", err)
	}
	fmt.Printf("   Input: 'AgenticGoKit'\n")
	fmt.Printf("   Operation: reverse\n")
	fmt.Printf("   Result: %v\n", textResult.Content)
	fmt.Printf("   Success: %v\n\n", textResult.Success)

	return nil
}

// demoMCPTools demonstrates using MCP server tools
func demoMCPTools() error {
	fmt.Println("🌐 MCP Tools Demo")
	fmt.Println("Testing MCP server tools...\n")

	ctx := context.Background()

	// Configure MCP server (example - adjust to your actual MCP server)
	server := vnext.MCPServer{
		Name:    "blog-http-sse",
		Type:    "http_sse",
		Address: "localhost",
		Port:    8812,
		Enabled: true,
	}

	fmt.Printf("Connecting to MCP server: %s (%s:%d)\n", server.Name, server.Address, server.Port)

	// Note: This requires an actual MCP server running
	// If no server is running, this will show an error (which is expected)

	// Try to execute echo tool if available
	echoResult, err := vnext.ExecuteToolByName(ctx, "echo", map[string]interface{}{
		"message": "Hello from MCP!",
	})
	if err != nil {
		fmt.Printf("   ⚠️  Echo tool not available (this is normal if no MCP server is running)\n")
		fmt.Printf("   Error: %v\n\n", err)
	} else {
		fmt.Printf("   Echo result: %v\n", echoResult.Content)
		fmt.Printf("   Success: %v\n\n", echoResult.Success)
	}

	return nil
}

// demoAgentWithTools demonstrates an agent using both native and MCP tools
func demoAgentWithTools() error {
	fmt.Println("🤖 Agent with Tools Demo")
	fmt.Println("Creating an agent with access to all tools...\n")

	ctx := context.Background()

	// Optional: configure MCP servers if you have them running
	mcpServer := vnext.MCPServer{
		Name:    "blog-http-sse",
		Type:    "http_sse",
		Address: "localhost",
		Port:    8812,
		Enabled: true,
	}

	// Create agent with custom handler that uses tools
	agent, err := vnext.NewBuilder("multi-tool-agent").
		WithConfig(&vnext.Config{
			Name: "multi-tool-agent",
			SystemPrompt: `You are a helpful assistant with access to various tools.
You can use:
- calculator: for mathematical operations
- get_timestamp: to get current time
- text_processor: to manipulate text
Use tools when they can help answer the user's question.`,
			Timeout: 60 * time.Second,
			LLM: vnext.LLMConfig{
				Provider:    "ollama",
				Model:       "gemma3:1b",
				Temperature: 0.7,
				MaxTokens:   300,
			},
		}).
		WithTools(
			vnext.WithMCP(mcpServer), // Add MCP server
			vnext.WithToolTimeout(30*time.Second),
		).
		WithHandler(func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
			// Custom logic: check if input needs tools
			if caps.Tools == nil {
				return caps.LLM("You are a helpful assistant.", input)
			}

			// Example: if input mentions calculation
			if strings.Contains(strings.ToLower(input), "calculate") ||
				strings.Contains(strings.ToLower(input), "compute") {
				// Use calculator tool
				if caps.Tools.IsAvailable("calculator") {
					result, err := caps.Tools.Execute(ctx, "calculator", map[string]interface{}{
						"operation": "add",
						"a":         42.0,
						"b":         13.0,
					})
					if err == nil && result.Success {
						return fmt.Sprintf("I used the calculator tool. Result: %v", result.Content), nil
					}
				}
			}

			// Example: if input asks for time
			if strings.Contains(strings.ToLower(input), "time") ||
				strings.Contains(strings.ToLower(input), "timestamp") {
				if caps.Tools.IsAvailable("get_timestamp") {
					result, err := caps.Tools.Execute(ctx, "get_timestamp", map[string]interface{}{
						"format": "human",
					})
					if err == nil && result.Success {
						return fmt.Sprintf("Current time: %v", result.Content), nil
					}
				}
			}

			// For other queries, use LLM
			return caps.LLM(
				"You are a helpful assistant with access to tools.",
				input,
			)
		}).
		Build()

	if err != nil {
		return fmt.Errorf("failed to build agent: %w", err)
	}

	// Test queries
	queries := []string{
		"What tools do you have access to?",
		"Can you calculate something for me?",
		"What time is it?",
	}

	for i, query := range queries {
		fmt.Printf("%d. Query: %s\n", i+1, query)
		result, err := agent.Run(ctx, query)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n\n", err)
			continue
		}
		fmt.Printf("   💬 Response: %s\n", result.Content)
		fmt.Printf("   ⏱️  Duration: %v\n\n", result.Duration)
	}

	return nil
}
