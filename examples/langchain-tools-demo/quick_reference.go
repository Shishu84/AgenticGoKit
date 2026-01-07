package main

import (
	"context"
	"fmt"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// QUICK REFERENCE: How to Add Tools to AgenticGoKit
// ================================================

// ============================================================================
// 1. NATIVE GO TOOL - Simple Example
// ============================================================================

type SimpleGreeterTool struct{}

func (t *SimpleGreeterTool) Name() string {
	return "greeter"
}

func (t *SimpleGreeterTool) Description() string {
	return "Greets a person by name. Args: name (string)"
}

func (t *SimpleGreeterTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	name, ok := args["name"].(string)
	if !ok {
		return &vnext.ToolResult{Success: false, Error: "name required"}, fmt.Errorf("name required")
	}

	return &vnext.ToolResult{
		Success: true,
		Content: map[string]interface{}{
			"greeting": fmt.Sprintf("Hello, %s! 👋", name),
		},
	}, nil
}

// ============================================================================
// 2. REGISTER YOUR TOOL - Call in init()
// ============================================================================

func init() {
	// Register the greeter tool
	vnext.RegisterInternalTool("greeter", func() vnext.Tool {
		return &SimpleGreeterTool{}
	})
}

// ============================================================================
// 3. USE THE TOOL - Three Ways
// ============================================================================

func ExampleDirectExecution() {
	ctx := context.Background()

	// Execute tool directly
	result, err := vnext.ExecuteToolByName(ctx, "greeter", map[string]interface{}{
		"name": "Alice",
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Result: %v\n", result.Content)
	// Output: Result: map[greeting:Hello, Alice! 👋]
}

func ExampleInHandler() {
	// Use in agent handler
	_, _ = vnext.NewBuilder("agent").
		WithHandler(func(ctx context.Context, input string, caps *vnext.Capabilities) (string, error) {
			if caps.Tools != nil && caps.Tools.IsAvailable("greeter") {
				result, _ := caps.Tools.Execute(ctx, "greeter", map[string]interface{}{
					"name": input,
				})
				return fmt.Sprintf("%v", result.Content), nil
			}
			return caps.LLM("You are helpful.", input)
		}).
		Build()
}

func ExampleToolDiscovery() {
	// Discover all available tools
	tools, err := vnext.DiscoverTools()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, tool := range tools {
		fmt.Printf("- %s: %s\n", tool.Name(), tool.Description())
	}
}

// ============================================================================
// 4. ADD MCP SERVER TOOLS
// ============================================================================

func ExampleMCPServers() {
	// Method 1: Explicit server list
	servers := []vnext.MCPServer{
		{Name: "filesystem", Type: "stdio", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-filesystem"}, Enabled: true},
		{Name: "web-api", Type: "http_sse", Address: "localhost", Port: 8080, Enabled: true},
	}

	_, _ = vnext.NewBuilder("agent").
		WithTools(vnext.WithMCP(servers...)).
		Build()

	// Method 2: Auto-discovery
	_, _ = vnext.NewBuilder("agent").
		WithTools(vnext.WithMCPDiscovery(8080, 8081)).
		Build()
}

// ============================================================================
// 5. TEMPLATE FOR YOUR CUSTOM TOOL
// ============================================================================

// Copy this template to create your own tool:

/*
type MyCustomTool struct{
    // Add any fields you need (API keys, config, etc.)
}

func (t *MyCustomTool) Name() string {
    return "my_tool_name"  // Must be unique
}

func (t *MyCustomTool) Description() string {
    return "What your tool does. Args: param1 (type), param2 (type)"
}

func (t *MyCustomTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
    // 1. Extract arguments
    param1, ok := args["param1"].(string)
    if !ok {
        return &vnext.ToolResult{Success: false, Error: "param1 required"}, fmt.Errorf("param1 required")
    }

    // 2. Do your logic
    result := doSomething(param1)

    // 3. Return result
    return &vnext.ToolResult{
        Success: true,
        Content: map[string]interface{}{
            "result": result,
        },
    }, nil
}

// Register it
func init() {
    vnext.RegisterInternalTool("my_tool_name", func() vnext.Tool {
        return &MyCustomTool{}
    })
}
*/

// ============================================================================
// 6. COMMON PATTERNS
// ============================================================================

// Pattern 1: Tool with validation
func toolWithValidation(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	// Validate required args
	required := []string{"param1", "param2"}
	for _, key := range required {
		if _, ok := args[key]; !ok {
			return &vnext.ToolResult{
				Success: false,
				Error:   fmt.Sprintf("missing required argument: %s", key),
			}, fmt.Errorf("missing %s", key)
		}
	}

	// Process...
	return &vnext.ToolResult{Success: true}, nil
}

// Pattern 2: Tool with error handling
func toolWithErrorHandling(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	result, err := someOperation(args)
	if err != nil {
		return &vnext.ToolResult{
			Success: false,
			Error:   err.Error(),
			Content: map[string]interface{}{
				"error_details": err.Error(),
			},
		}, err
	}

	return &vnext.ToolResult{
		Success: true,
		Content: map[string]interface{}{
			"result": result,
		},
	}, nil
}

// Pattern 3: Tool with context cancellation
func toolWithContext(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return &vnext.ToolResult{
			Success: false,
			Error:   "operation cancelled",
		}, ctx.Err()
	default:
		// Continue processing
	}

	// Do work...
	return &vnext.ToolResult{Success: true}, nil
}

// Dummy helper functions for examples
func someOperation(args map[string]interface{}) (string, error) {
	return "success", nil
}
