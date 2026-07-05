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

## Scope

- Uses Anthropic's provider-side token counting endpoint.
- Supports structured message requests, tools, system prompts, and extra request fields.
- Does not provide local token IDs.
- Does not claim billing parity; final response usage remains authoritative.
