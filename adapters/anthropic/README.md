# Anthropic Adapter

API-backed token counting for Anthropic Messages.

There is no supported local tokenizer for current Claude models, so this adapter calls Anthropic's server-side count endpoint instead of pretending to provide local parity.

```powershell
go get github.com/ron2111/omnitoken/adapters/anthropic
```

## Usage

```go
client := anthropic.Client{APIKey: apiKey}

result, err := client.CountMessageTokens(ctx, anthropic.CountRequest{
	Model: "claude-sonnet-4-5",
	Messages: []anthropic.Message{
		{Role: "user", Content: "hello"},
	},
})
if err != nil {
	panic(err)
}

fmt.Println(result.InputTokens)
```

Parse final response usage when reconciling actual provider usage:

```go
usage, err := anthropic.ParseUsage(responseJSON)
if err != nil {
	panic(err)
}

fmt.Println(usage.InputTokens, usage.OutputTokens, usage.CacheReadInputTokens)
```

## Scope

- Uses Anthropic's provider-side token counting endpoint.
- Supports structured message requests, tools, system prompts, and extra request fields.
- Parses final Messages API usage blocks for post-execution reconciliation.
- Does not provide local token IDs.
- Does not claim billing parity; final response usage remains authoritative.
