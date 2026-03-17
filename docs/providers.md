# Providers

A **Provider** is the abstraction that connects the SDK to an AI backend. It handles HTTP communication, request/response mapping, and streaming protocol details.

## The Provider Interface

```go
type Provider interface {
    Name() string
    GetModels() ([]Model, error)
    DoGenerate(ctx context.Context, params GenerateParams) (*GenerateResult, error)
    DoStream(ctx context.Context, params GenerateParams) (*StreamResult, error)
}
```

| Method | Purpose |
|--------|---------|
| `Name()` | Returns a human-readable provider identifier (e.g. `"openai-completions"`) |
| `GetModels()` | Lists available models (optional, may return nil) |
| `DoGenerate()` | Performs a single non-streaming LLM call |
| `DoStream()` | Performs a streaming LLM call, returning a channel of `StreamPart` |

The SDK never calls a provider directly — it goes through the `Client` which adds orchestration (tool loop, callbacks, multi-step). The `Model` struct carries a reference to its provider:

```go
type Model struct {
    ID          string
    DisplayName string
    Provider    Provider
    Type        ModelType   // "chat"
    MaxTokens   int
}
```

## OpenAI Completions Provider

The `provider/openai/completions` package provides an implementation for the OpenAI Chat Completions API (`/chat/completions`).

### Basic Usage

```go
import "github.com/memohai/twilight-ai/provider/openai/completions"

provider := completions.New(
    completions.WithAPIKey("sk-..."),
)
model := provider.ChatModel("gpt-4o-mini")
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithAPIKey(key)` | `""` | API key sent as `Authorization: Bearer <key>` |
| `WithBaseURL(url)` | `https://api.openai.com/v1` | Base URL for API requests |
| `WithHTTPClient(client)` | `&http.Client{}` | Custom HTTP client (for proxies, timeouts, etc.) |

### OpenAI-Compatible Providers

Any service that implements the OpenAI Chat Completions API works out of the box:

```go
// DeepSeek
provider := completions.New(
    completions.WithAPIKey("your-deepseek-key"),
    completions.WithBaseURL("https://api.deepseek.com"),
)

// Groq
provider := completions.New(
    completions.WithAPIKey("your-groq-key"),
    completions.WithBaseURL("https://api.groq.com/openai/v1"),
)

// Azure OpenAI
provider := completions.New(
    completions.WithAPIKey("your-azure-key"),
    completions.WithBaseURL("https://your-resource.openai.azure.com/openai/deployments/gpt-4o"),
)

// Local (Ollama, vLLM, etc.)
provider := completions.New(
    completions.WithBaseURL("http://localhost:11434/v1"),
)
```

### Supported Features

| Feature | Supported |
|---------|-----------|
| Chat completions | ✅ |
| Streaming (SSE) | ✅ |
| Tool/function calling | ✅ |
| Vision (image inputs) | ✅ |
| Reasoning content (o1, DeepSeek-R1) | ✅ |
| JSON mode / JSON Schema | ✅ |
| Token usage reporting | ✅ |
| Cached token details | ✅ |

### Custom HTTP Client

Use `WithHTTPClient` for custom timeouts, proxies, or TLS settings:

```go
provider := completions.New(
    completions.WithAPIKey("sk-..."),
    completions.WithHTTPClient(&http.Client{
        Timeout: 120 * time.Second,
        Transport: &http.Transport{
            Proxy: http.ProxyFromEnvironment,
        },
    }),
)
```

## OpenAI Responses Provider

The `provider/openai/responses` package provides an implementation for the OpenAI Responses API (`/responses`). This is OpenAI's newer API that offers first-class reasoning support, URL citation annotations, and a flat input format.

### When to Use Responses vs Completions

| | Chat Completions | Responses |
|--|---|---|
| **Endpoint** | `/chat/completions` | `/responses` |
| **Reasoning models** | Basic support (`reasoning_content` field) | First-class (`reasoning` output items with summaries) |
| **Citations** | Not supported | URL citations via annotations |
| **Input format** | Nested `messages` array | Flat `input` array |
| **Compatibility** | Broad (DeepSeek, Groq, Ollama, etc.) | OpenAI and OpenRouter |

Use **Completions** when you need broad compatibility with OpenAI-compatible endpoints. Use **Responses** when you want native reasoning model support (o3, o4-mini) or URL citation annotations.

### Basic Usage

```go
import "github.com/memohai/twilight-ai/provider/openai/responses"

provider := responses.New(
    responses.WithAPIKey("sk-..."),
)
model := provider.ChatModel("gpt-4o-mini")
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithAPIKey(key)` | `""` | API key sent as `Authorization: Bearer <key>` |
| `WithBaseURL(url)` | `https://api.openai.com/v1` | Base URL for API requests |
| `WithHTTPClient(client)` | `&http.Client{}` | Custom HTTP client |

### Using with OpenRouter

OpenRouter supports the Responses API as a beta feature:

```go
provider := responses.New(
    responses.WithAPIKey("sk-or-v1-..."),
    responses.WithBaseURL("https://openrouter.ai/api/v1"),
)
model := provider.ChatModel("openai/o4-mini")
```

### Reasoning Models

Reasoning models (o3, o4-mini) return both reasoning summaries and the final answer:

```go
effort := "medium"
result, _ := sdk.GenerateTextResult(ctx,
    sdk.WithModel(provider.ChatModel("openai/o4-mini")),
    sdk.WithMessages([]sdk.Message{
        sdk.UserMessage("What is 15 * 37? Think step by step."),
    }),
    sdk.WithReasoningEffort(&effort),
)
fmt.Println(result.Reasoning)  // model's reasoning summary
fmt.Println(result.Text)       // final answer: "555"
```

In streaming mode, reasoning arrives as `ReasoningStartPart` / `ReasoningDeltaPart` / `ReasoningEndPart` before the text content.

### Supported Features

| Feature | Supported |
|---------|-----------|
| Text generation | ✅ |
| Streaming (SSE) | ✅ |
| Tool/function calling | ✅ |
| Vision (image inputs) | ✅ |
| Reasoning summaries (o3, o4-mini) | ✅ |
| URL citation annotations | ✅ |
| JSON mode / JSON Schema | ✅ |
| Token usage reporting | ✅ |
| Cached / reasoning token details | ✅ |

## Anthropic Provider

The `provider/anthropic/messages` package implements the [Anthropic Messages API](https://docs.anthropic.com/en/api/messages) for Claude models.

### Basic Usage

```go
import "github.com/memohai/twilight-ai/provider/anthropic/messages"

provider := messages.New(
    messages.WithAPIKey("sk-ant-..."),
)
model := provider.ChatModel("claude-sonnet-4-20250514")
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithAPIKey(key)` | `""` | API key sent as `x-api-key` header |
| `WithAuthToken(token)` | `""` | OAuth token sent as `Authorization: Bearer <token>` |
| `WithBaseURL(url)` | `https://api.anthropic.com` | Base URL for API requests |
| `WithHTTPClient(client)` | `&http.Client{}` | Custom HTTP client |
| `WithThinking(config)` | `nil` | Enable extended thinking for reasoning |

### Extended Thinking

Claude supports [extended thinking](https://docs.anthropic.com/en/docs/build-with-claude/extended-thinking) (chain-of-thought reasoning):

```go
provider := messages.New(
    messages.WithAPIKey("sk-ant-..."),
    messages.WithThinking(messages.ThinkingConfig{
        Type:         "enabled",
        BudgetTokens: 10000,
    }),
)
```

When enabled, the model's internal reasoning appears in `result.Reasoning` (non-streaming) or as `ReasoningStartPart` / `ReasoningDeltaPart` / `ReasoningEndPart` events (streaming).

### Supported Features

| Feature | Supported |
|---------|-----------|
| Text generation | ✅ |
| Streaming (SSE) | ✅ |
| Tool/function calling | ✅ |
| Vision (image inputs) | ✅ |
| Extended thinking | ✅ |
| Token usage reporting | ✅ |
| Cached token details | ✅ |

---

## Google Gemini Provider

The `provider/google/generativeai` package implements the [Google Generative AI API](https://ai.google.dev/api) for Gemini models.

### Basic Usage

```go
import "github.com/memohai/twilight-ai/provider/google/generativeai"

provider := generativeai.New(
    generativeai.WithAPIKey("AIza..."),
)
model := provider.ChatModel("gemini-2.5-flash")
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithAPIKey(key)` | `""` | API key sent as `x-goog-api-key` header |
| `WithBaseURL(url)` | `https://generativelanguage.googleapis.com/v1beta` | Base URL |
| `WithHTTPClient(client)` | `&http.Client{}` | Custom HTTP client |

### Model ID

The model ID can be a simple name or a full resource path:

```go
// Simple name — resolved to "models/gemini-2.5-flash"
model := provider.ChatModel("gemini-2.5-flash")

// Full path — used as-is
model := provider.ChatModel("publishers/google/models/gemini-2.5-flash")
```

### API Endpoints

| Operation | Endpoint |
|-----------|----------|
| Non-streaming | `POST {baseURL}/models/{modelId}:generateContent` |
| Streaming | `POST {baseURL}/models/{modelId}:streamGenerateContent?alt=sse` |

### How Messages Are Mapped

The provider automatically converts SDK messages to Google's format:

| SDK | Google API |
|-----|-----------|
| `System` param | `systemInstruction` field (separate from `contents`) |
| User message | `{role: "user", parts: [{text: "..."}, ...]}` |
| Assistant message | `{role: "model", parts: [{text: "..."}, {functionCall: ...}]}` |
| Tool result message | `{role: "user", parts: [{functionResponse: {name, response}}]}` |

### Tool Choice Mapping

| SDK `ToolChoice` | Google `functionCallingConfig.mode` |
|------------------|-------------------------------------|
| `"auto"` | `AUTO` |
| `"none"` | `NONE` |
| `"required"` | `ANY` |

### Thinking / Reasoning

Gemini 2.5+ models support thinking (reasoning). The model returns parts with `thought: true` which the provider maps to `Reasoning` in the result:

```go
provider := generativeai.New(generativeai.WithAPIKey("AIza..."))
model := provider.ChatModel("gemini-2.5-flash")

result, _ := sdk.GenerateTextResult(ctx,
    sdk.WithModel(model),
    sdk.WithMessages([]sdk.Message{
        sdk.UserMessage("What is 15 * 37? Think step by step."),
    }),
)
fmt.Println(result.Reasoning) // model's thinking process
fmt.Println(result.Text)      // final answer
```

### Supported Features

| Feature | Supported |
|---------|-----------|
| Text generation | ✅ |
| Streaming (SSE) | ✅ |
| Tool/function calling | ✅ |
| Vision (image inputs) | ✅ |
| Thinking / Reasoning (Gemini 2.5+) | ✅ |
| JSON mode | ✅ |
| Token usage reporting | ✅ |
| Cached content token details | ✅ |

---

## Implementing a Custom Provider

To add support for a new AI backend, implement the `sdk.Provider` interface:

```go
package myprovider

import (
    "context"
    "github.com/memohai/twilight-ai/sdk"
)

type MyProvider struct {
    apiKey string
}

func New(apiKey string) *MyProvider {
    return &MyProvider{apiKey: apiKey}
}

func (p *MyProvider) Name() string {
    return "my-provider"
}

func (p *MyProvider) GetModels() ([]sdk.Model, error) {
    return []sdk.Model{
        {ID: "my-model-v1", Provider: p, Type: sdk.ModelTypeChat},
    }, nil
}

func (p *MyProvider) ChatModel(id string) *sdk.Model {
    return &sdk.Model{ID: id, Provider: p, Type: sdk.ModelTypeChat}
}

func (p *MyProvider) DoGenerate(ctx context.Context, params sdk.GenerateParams) (*sdk.GenerateResult, error) {
    // Make HTTP request to your backend...
    // Map the response to *sdk.GenerateResult
    return &sdk.GenerateResult{
        Text:         "response text",
        FinishReason: sdk.FinishReasonStop,
    }, nil
}

func (p *MyProvider) DoStream(ctx context.Context, params sdk.GenerateParams) (*sdk.StreamResult, error) {
    ch := make(chan sdk.StreamPart, 64)

    go func() {
        defer close(ch)
        // Stream chunks from your backend...
        ch <- &sdk.StartPart{}
        ch <- &sdk.StartStepPart{}
        ch <- &sdk.TextStartPart{}
        ch <- &sdk.TextDeltaPart{Text: "Hello"}
        ch <- &sdk.TextEndPart{}
        ch <- &sdk.FinishStepPart{FinishReason: sdk.FinishReasonStop}
        ch <- &sdk.FinishPart{FinishReason: sdk.FinishReasonStop}
    }()

    return &sdk.StreamResult{Stream: ch}, nil
}
```

Then use it exactly like the built-in provider:

```go
provider := myprovider.New("my-key")
model := provider.ChatModel("my-model-v1")

text, err := sdk.GenerateText(ctx,
    sdk.WithModel(model),
    sdk.WithMessages([]sdk.Message{sdk.UserMessage("Hello")}),
)
```

## Next Steps

- [Tool Calling](tools.md) — define tools and enable multi-step execution
- [Streaming](streaming.md) — understand StreamPart types
- [API Reference](api-reference.md) — complete type and function reference
