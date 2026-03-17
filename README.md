# Twilight AI

A lightweight, idiomatic AI SDK for Go — inspired by [Vercel AI SDK](https://sdk.vercel.ai/).

[![Go Reference](https://pkg.go.dev/badge/github.com/memohai/twilight-ai.svg)](https://pkg.go.dev/github.com/memohai/twilight-ai)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

## Features

- **Simple API** — `GenerateText` and `StreamText`, two functions cover most use cases
- **Provider-agnostic** — swap between OpenAI, Azure, or any OpenAI-compatible endpoint
- **Tool calling** — define tools with Go structs, SDK infers JSON Schema and handles multi-step execution
- **Streaming** — first-class channel-based streaming with fine-grained `StreamPart` types
- **Multi-step execution** — automatic tool-call loop with configurable `MaxSteps`
- **Rich message types** — text, images, files, reasoning content, tool calls/results
- **Approval flow** — optional human-in-the-loop approval for sensitive tool calls
- **Minimal dependencies** — only [google/jsonschema-go](https://github.com/google/jsonschema-go) beyond the standard library

## Installation

```bash
go get github.com/memohai/twilight-ai
```

Requires **Go 1.25+**.

## Quick Start

### Generate Text

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/memohai/twilight-ai/provider/openai"
    "github.com/memohai/twilight-ai/sdk"
)

func main() {
    provider := openai.NewCompletions(
        openai.WithAPIKey("sk-..."),
    )
    model := provider.ChatModel("gpt-4o-mini")

    text, err := sdk.GenerateText(context.Background(),
        sdk.WithModel(model),
        sdk.WithMessages([]sdk.Message{
            sdk.UserMessage("Explain Go channels in 3 sentences."),
        }),
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(text)
}
```

### Stream Text

```go
sr, err := sdk.StreamText(ctx,
    sdk.WithModel(model),
    sdk.WithMessages([]sdk.Message{
        sdk.UserMessage("Write a haiku about concurrency."),
    }),
)
if err != nil {
    log.Fatal(err)
}

for part := range sr.Stream {
    switch p := part.(type) {
    case *sdk.TextDeltaPart:
        fmt.Print(p.Text)
    case *sdk.ErrorPart:
        log.Fatal(p.Error)
    }
}
```

### Tool Calling

Define a struct for your tool's parameters — the SDK infers the JSON Schema automatically:

```go
type WeatherParams struct {
    City string `json:"city" jsonschema:"City name"`
}

weatherTool := sdk.NewTool("get_weather", "Get current weather for a city",
    func(ctx *sdk.ToolExecContext, input WeatherParams) (any, error) {
        return map[string]any{"city": input.City, "temp": "22°C"}, nil
    },
)

result, err := sdk.GenerateTextResult(ctx,
    sdk.WithModel(model),
    sdk.WithMessages([]sdk.Message{
        sdk.UserMessage("What's the weather in Tokyo?"),
    }),
    sdk.WithTools([]sdk.Tool{weatherTool}),
    sdk.WithMaxSteps(5),
)
```

## Architecture

```
┌──────────────────────────────────────────────┐
│                  Your App                     │
├──────────────────────────────────────────────┤
│  sdk.GenerateText / sdk.StreamText           │
│  ┌─────────────────────────────────────────┐ │
│  │  Client (orchestration, tool loop)      │ │
│  └─────────────┬───────────────────────────┘ │
│                │                             │
│  ┌─────────────▼───────────────────────────┐ │
│  │  Provider interface                     │ │
│  │  DoGenerate() / DoStream()              │ │
│  └─────────────┬───────────────────────────┘ │
├────────────────┼─────────────────────────────┤
│  ┌─────────────▼──┐  ┌──────────────────┐   │
│  │  OpenAI        │  │  Your Provider   │   │
│  │  Completions   │  │  (coming soon)   │   │
│  └────────────────┘  └──────────────────┘   │
└──────────────────────────────────────────────┘
```

## Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/getting-started.md) | Installation, setup, and first request |
| [Providers](docs/providers.md) | Provider interface and OpenAI implementation |
| [Tool Calling](docs/tools.md) | Defining tools, multi-step execution, approval flow |
| [Streaming](docs/streaming.md) | Channel-based streaming and StreamPart types |
| [API Reference](docs/api-reference.md) | Complete type and function reference |

## Supported Providers

| Provider | Package | Status |
|----------|---------|--------|
| OpenAI | `provider/openai` | ✅ Stable |
| OpenAI-compatible (DeepSeek, Groq, etc.) | `provider/openai` + `WithBaseURL` | ✅ Stable |
| Anthropic | — | Planned |
| Google Gemini | — | Planned |

## License

[Apache License 2.0](LICENSE)
