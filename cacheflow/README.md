# Cacheflow

`cacheflow` is OmniToken's dependency-free prompt-cache analysis package.

It is built for Go teams that want to understand whether rendered prompts have stable token prefixes before sending them to OpenAI, Anthropic, Gemini, or another provider. It does not claim billing parity and it does not call provider APIs.

## What It Does

- Counts prompt tokens with any `omnitoken.ModelEngine`.
- Calculates cache-boundary alignment for one prompt.
- Reads JSONL prompt traces.
- Finds common token prefixes across repeated prompts.
- Estimates reusable prefix tokens under a local cache profile.
- Emits best-effort cache-breaker hints for timestamps, UUIDs, request IDs, and dynamic metadata.

## What It Does Not Do

- It does not edit prompts automatically.
- It does not guarantee provider billing behavior.
- It does not require network calls or credentials.
- It does not add dependencies to the root module.

## Align One Prompt

```go
engine, err := omnitoken.ForModel("gpt-4o")
if err != nil {
	panic(err)
}

report := cacheflow.NewAligner(engine).AlignPromptToProfile(systemPrompt, cacheflow.ProfileOpenAI)
fmt.Println(report.CurrentTokens, report.PaddingNeeded)
```

## Simulate A Trace

```go
items := []cacheflow.TraceItem{
	{ID: "1", Prompt: stablePrefix + "user question one"},
	{ID: "2", Prompt: stablePrefix + "user question two"},
}

report := cacheflow.Simulate(engine, items, cacheflow.SimulationOptions{
	Profile:        cacheflow.ProfileOpenAI,
	DetectBreakers: true,
})
fmt.Println(report.ReusablePrefixTokens)
```

## JSONL Format

Raw rendered prompts:

```json
{"id":"1","model":"gpt-4o","prompt":"..."}
{"id":"2","model":"gpt-4o","prompt":"..."}
```

Structured prompt parts:

```json
{"id":"1","model":"gpt-4o","parts":[{"name":"system","stable":true,"text":"..."},{"name":"user","stable":false,"text":"..."}]}
```

For structured parts, `cacheflow` concatenates `parts[].text` in order and uses `stable` only for diagnostics.

## CLI

```powershell
omni cache -model gpt-4o -profile openai "hello world"
omni cache-sim -model gpt-4o -profile openai -input prompts.jsonl -breakers
```

## Profiles

```go
cacheflow.ProfileGeneric
cacheflow.ProfileOpenAI
```

Profiles are local planning helpers. Provider behavior can change, and final usage metadata remains authoritative.
