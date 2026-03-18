# Twilight AI

A lightweight, idiomatic AI SDK for Go — inspired by [Vercel AI SDK](https://sdk.vercel.ai/).

[![Go Reference](https://pkg.go.dev/badge/github.com/memohai/twilight-ai.svg)](https://pkg.go.dev/github.com/memohai/twilight-ai)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

## Features

- **Simple API** — `GenerateText`, `StreamText`, `Embed`, and `EmbedMany` cover most use cases
- **Provider-agnostic** — swap between OpenAI, Anthropic, Google, or any OpenAI-compatible endpoint
- **Model discovery** — `ListModels` fetches available models, `Test` checks provider connectivity and model support
- **Tool calling** — define tools with Go structs, SDK infers JSON Schema and handles multi-step execution
- **Streaming** — first-class channel-based streaming with fine-grained `StreamPart` types
- **Multi-step execution** — automatic tool-call loop with configurable `MaxSteps`
- **Rich message types** — text, images, files, reasoning content, tool calls/results
- **Embeddings** — generate embeddings with `Embed` / `EmbedMany`, supports OpenAI and Google providers
- **Approval flow** — optional human-in-the-loop approval for sensitive tool calls
- **Minimal dependencies** — only [google/jsonschema-go](https://github.com/google/jsonschema-go) beyond the standard library

## Installation

```bash
go get github.com/memohai/twilight-ai
```

Requires **Go 1.25+**.

## Quick Start

### Generate Text (Chat Completions API)

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/memohai/twilight-ai/provider/openai/completions"
    "github.com/memohai/twilight-ai/sdk"
)

func main() {
    provider := completions.New(
        completions.WithAPIKey("sk-..."),
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

### Generate Text (Responses API)

```go
import "github.com/memohai/twilight-ai/provider/openai/responses"

provider := responses.New(
    responses.WithAPIKey("sk-..."),
)
model := provider.ChatModel("gpt-4o-mini")

text, err := sdk.GenerateText(context.Background(),
    sdk.WithModel(model),
    sdk.WithMessages([]sdk.Message{
        sdk.UserMessage("Explain Go channels in 3 sentences."),
    }),
)
```

The Responses API is OpenAI's newer API with first-class support for reasoning models (o3, o4-mini), URL citation annotations, and a flat input format. See [Providers](docs/providers.md) for details.

### Anthropic

```go
import "github.com/memohai/twilight-ai/provider/anthropic/messages"

provider := messages.New(
    messages.WithAPIKey("sk-ant-..."),
)
model := provider.ChatModel("claude-sonnet-4-20250514")

maxTokens := 1024
text, err := sdk.GenerateText(context.Background(),
    sdk.WithModel(model),
    sdk.WithMaxTokens(maxTokens),
    sdk.WithMessages([]sdk.Message{
        sdk.UserMessage("Explain Go channels in 3 sentences."),
    }),
)
```

For extended thinking (reasoning), configure the provider with `WithThinking`:

```go
provider := messages.New(
    messages.WithAPIKey("sk-ant-..."),
    messages.WithThinking(messages.ThinkingConfig{
        Type:         "enabled",
        BudgetTokens: 4000,
    }),
)
```

### Google Gemini

```go
import "github.com/memohai/twilight-ai/provider/google/generativeai"

provider := generativeai.New(
    generativeai.WithAPIKey("AIza..."),
)
model := provider.ChatModel("gemini-2.5-flash")

text, err := sdk.GenerateText(context.Background(),
    sdk.WithModel(model),
    sdk.WithMessages([]sdk.Message{
        sdk.UserMessage("Explain Go channels in 3 sentences."),
    }),
)
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

### Embeddings

Generate vector embeddings for text using OpenAI or Google:

```go
import "github.com/memohai/twilight-ai/provider/openai/embedding"

provider := embedding.New(embedding.WithAPIKey("sk-..."))
model := provider.EmbeddingModel("text-embedding-3-small")

// Single value
vec, err := sdk.Embed(ctx, "Hello world", sdk.WithEmbeddingModel(model))
// vec is []float64

// Multiple values
result, err := sdk.EmbedMany(ctx, []string{"Hello", "World"},
    sdk.WithEmbeddingModel(model),
    sdk.WithDimensions(256),
)
// result.Embeddings is [][]float64
// result.Usage.Tokens reports token consumption
```

Google Gemini embeddings:

```go
import "github.com/memohai/twilight-ai/provider/google/embedding"

provider := embedding.New(
    embedding.WithAPIKey("AIza..."),
    embedding.WithTaskType("RETRIEVAL_DOCUMENT"),
)
model := provider.EmbeddingModel("gemini-embedding-001")

vec, err := sdk.Embed(ctx, "Hello world", sdk.WithEmbeddingModel(model))
```

### Provider Health Check & Model Discovery

Test connectivity and discover available models before making generation requests:

```go
provider := completions.New(completions.WithAPIKey("sk-..."))

// Check provider connectivity
result := provider.Test(context.Background())
switch result.Status {
case sdk.ProviderStatusOK:
    fmt.Println("Provider is healthy")
case sdk.ProviderStatusUnhealthy:
    fmt.Println("Connected but unhealthy:", result.Message)
case sdk.ProviderStatusUnreachable:
    fmt.Println("Cannot connect:", result.Message)
}

// List all available models
models, err := provider.ListModels(context.Background())
for _, m := range models {
    fmt.Println(m.ID)
}

// Check if a specific model is supported
model := provider.ChatModel("gpt-4o")
testResult, err := model.Test(context.Background())
if testResult.Supported {
    fmt.Println("Model is supported")
}
```

## Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/getting-started.md) | Installation, setup, and first request |
| [Providers](docs/providers.md) | Provider interface, OpenAI, Anthropic, and Google Gemini |
| [Embeddings](docs/embeddings.md) | Generate vector embeddings with OpenAI and Google |
| [Tool Calling](docs/tools.md) | Defining tools, multi-step execution, approval flow |
| [Streaming](docs/streaming.md) | Channel-based streaming and StreamPart types |
| [API Reference](docs/api-reference.md) | Complete type and function reference |

## Supported Providers

| Provider | Constructor | API | Status |
|----------|-------------|-----|--------|
| OpenAI Chat Completions | `completions.New()` | `/chat/completions` | ✅ Stable |
| OpenAI Responses | `responses.New()` | `/responses` | ✅ Stable |
| OpenAI-compatible (DeepSeek, Groq, etc.) | `completions.New()` + `WithBaseURL` | `/chat/completions` | ✅ Stable |
| OpenRouter Responses | `responses.New()` + `WithBaseURL` | `/responses` | ✅ Stable |
| Anthropic | `messages.New()` | `/messages` | ✅ Stable |
| Google Gemini | `generativeai.New()` | Generative AI API | ✅ Stable |
| OpenAI Embeddings | `embedding.New()` | `/embeddings` | ✅ Stable |
| Google Embeddings | `embedding.New()` | `embedContent` / `batchEmbedContents` | ✅ Stable |

## License

[Apache License 2.0](LICENSE)
