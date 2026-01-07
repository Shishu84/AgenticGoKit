package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// CalculatorTool performs mathematical operations
type CalculatorTool struct{}

func (t *CalculatorTool) Name() string {
	return "calculator"
}

func (t *CalculatorTool) Description() string {
	return "Performs basic arithmetic operations: add, subtract, multiply, divide. Args: operation (string), a (number), b (number)"
}

func (t *CalculatorTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return &vnext.ToolResult{
			Success: false,
			Error:   "missing required argument: operation (string)",
		}, fmt.Errorf("operation required")
	}

	a, ok := args["a"].(float64)
	if !ok {
		return &vnext.ToolResult{
			Success: false,
			Error:   "missing required argument: a (number)",
		}, fmt.Errorf("a required")
	}

	b, ok := args["b"].(float64)
	if !ok {
		return &vnext.ToolResult{
			Success: false,
			Error:   "missing required argument: b (number)",
		}, fmt.Errorf("b required")
	}

	var result float64
	switch strings.ToLower(operation) {
	case "add", "addition", "+":
		result = a + b
	case "subtract", "subtraction", "-":
		result = a - b
	case "multiply", "multiplication", "*":
		result = a * b
	case "divide", "division", "/":
		if b == 0 {
			return &vnext.ToolResult{
				Success: false,
				Error:   "division by zero is not allowed",
			}, fmt.Errorf("division by zero")
		}
		result = a / b
	default:
		return &vnext.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported operation: %s. Use: add, subtract, multiply, divide", operation),
		}, fmt.Errorf("unsupported operation: %s", operation)
	}

	return &vnext.ToolResult{
		Success: true,
		Content: map[string]interface{}{
			"result":    result,
			"operation": operation,
			"a":         a,
			"b":         b,
		},
	}, nil
}

// TimestampTool returns current timestamp in various formats
type TimestampTool struct{}

func (t *TimestampTool) Name() string {
	return "get_timestamp"
}

func (t *TimestampTool) Description() string {
	return "Returns the current timestamp. Args: format (optional, string) - unix, rfc3339, or human (default: unix)"
}

func (t *TimestampTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	format := "unix"
	if f, ok := args["format"].(string); ok {
		format = strings.ToLower(f)
	}

	now := time.Now()
	var result interface{}

	switch format {
	case "unix":
		result = now.Unix()
	case "rfc3339":
		result = now.Format(time.RFC3339)
	case "human", "readable":
		result = now.Format("2006-01-02 15:04:05 MST")
	case "iso8601":
		result = now.Format(time.RFC3339)
	default:
		result = now.Unix()
	}

	return &vnext.ToolResult{
		Success: true,
		Content: map[string]interface{}{
			"timestamp": result,
			"format":    format,
			"timezone":  now.Location().String(),
		},
	}, nil
}

// TextProcessorTool performs text operations
type TextProcessorTool struct{}

func (t *TextProcessorTool) Name() string {
	return "text_processor"
}

func (t *TextProcessorTool) Description() string {
	return "Performs text operations. Args: operation (string: uppercase, lowercase, reverse, count), text (string)"
}

func (t *TextProcessorTool) Execute(ctx context.Context, args map[string]interface{}) (*vnext.ToolResult, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return &vnext.ToolResult{
			Success: false,
			Error:   "missing required argument: operation (string)",
		}, fmt.Errorf("operation required")
	}

	text, ok := args["text"].(string)
	if !ok {
		return &vnext.ToolResult{
			Success: false,
			Error:   "missing required argument: text (string)",
		}, fmt.Errorf("text required")
	}

	var result interface{}
	switch strings.ToLower(operation) {
	case "uppercase", "upper":
		result = strings.ToUpper(text)
	case "lowercase", "lower":
		result = strings.ToLower(text)
	case "reverse":
		runes := []rune(text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		result = string(runes)
	case "count", "length":
		result = len(text)
	case "words":
		result = len(strings.Fields(text))
	default:
		return &vnext.ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported operation: %s", operation),
		}, fmt.Errorf("unsupported operation: %s", operation)
	}

	return &vnext.ToolResult{
		Success: true,
		Content: map[string]interface{}{
			"result":    result,
			"operation": operation,
			"input":     text,
		},
	}, nil
}

// Register all native tools on package initialization
func init() {
	// Register calculator tool
	vnext.RegisterInternalTool("calculator", func() vnext.Tool {
		return &CalculatorTool{}
	})

	// Register timestamp tool
	vnext.RegisterInternalTool("get_timestamp", func() vnext.Tool {
		return &TimestampTool{}
	})

	// Register text processor tool
	vnext.RegisterInternalTool("text_processor", func() vnext.Tool {
		return &TextProcessorTool{}
	})
}
