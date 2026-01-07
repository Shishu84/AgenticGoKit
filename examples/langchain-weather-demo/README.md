# LangChain-Style Weather Tool (Automatic)

A minimal example showing LangChain-like simplicity in AgenticGoKit.

## Quick Start

```bash
cd examples/langchain-weather-demo
# Make sure ollama is running and model available
# ollama pull gemma3:1b
GO111MODULE=on go run .
```

## What it does
- Defines a single tool `check_weather`
- Registers it via `RegisterInternalTool`
- Builds an agent with `WithPreset(vnext.ChatAgent)` so tools are auto-used
- Runs a couple of queries; agent decides when to call the tool

## Code highlights
- Tool: `WeatherTool` implements `Name`, `Description`, `Execute`
- Registration: in `init()` with `RegisterInternalTool`
- Agent: built with `WithPreset` (no custom handler needed)

## Files
- `main.go` — everything in one file

## Change model/provider
Update in `main.go`:
```
LLM: vnext.LLMConfig{
    Provider: "ollama", // change if needed
    Model:    "gemma3:1b",
}
```
