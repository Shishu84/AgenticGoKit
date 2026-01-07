package main

import (
	"context"
	"fmt"
	"log"

	_ "github.com/agenticgokit/agenticgokit/plugins/llm/ollama"
	vnext "github.com/agenticgokit/agenticgokit/v1beta"
)

// SIMPLEST EXAMPLE - Tools work automatically!
// ============================================
// 1. Import custom_tools.go (has init() that registers tools)
// 2. Build agent with preset
// 3. Run - tools are automatically used!

func main() {
	ctx := context.Background()

	// That's it! Just create the agent with a preset
	agent, err := vnext.NewBuilder("auto-tool-agent").
		WithPreset(vnext.ChatAgent).
		Build()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("🤖 Agent with Automatic Tool Support")
	fmt.Println("=====================================\n")

	// Ask questions - agent automatically uses tools when needed!
	testQueries := []struct {
		query       string
		description string
	}{
		{
			query:       "Calculate 42 multiplied by 13",
			description: "Should use calculator tool",
		},
		{
			query:       "What time is it now?",
			description: "Should use timestamp tool",
		},
		{
			query:       "Reverse the text 'AgenticGoKit'",
			description: "Should use text_processor tool",
		},
		{
			query:       "What is 100 divided by 5?",
			description: "Should use calculator tool",
		},
		{
			query:       "Convert 'hello world' to uppercase",
			description: "Should use text_processor tool",
		},
	}

	for i, test := range testQueries {
		fmt.Printf("%d. %s\n", i+1, test.query)
		fmt.Printf("   ℹ️  %s\n", test.description)

		result, err := agent.Run(ctx, test.query)
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("   💬 Response: %s\n", result.Content)

		// Show which tools were automatically called
		if len(result.ToolsCalled) > 0 {
			fmt.Printf("   🔧 Tools used: %v\n", result.ToolsCalled)
		} else {
			fmt.Printf("   📝 No tools used (LLM answered directly)\n")
		}

		fmt.Printf("   ⏱️  Duration: %v\n\n", result.Duration)
	}

	fmt.Println("\n✅ Done! All tools were used automatically - no custom code needed!")
}
