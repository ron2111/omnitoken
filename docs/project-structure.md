# Project Structure

OmniToken intentionally keeps the root Go package flat. In Go, each directory is a separate package, so moving root files into folders would change import paths or force a larger internal-package refactor.

## Root Package

Root files implement `github.com/ron2111/omnitoken`:

- `engine.go`, `bpe.go`, `bpe_runtime.go`, `bpe_ranks.go`, `scanner.go`: OpenAI-compatible BPE runtime.
- `registry.go`: model and encoding registry.
- `options.go`, `special_tokens.go`: special-token-aware APIs.
- `wordpiece.go`, `sentencepiece.go`: lightweight custom tokenizer engines.
- `doc.go`, `example_test.go`: package documentation and examples.

Tests stay beside the package files because many validate unexported hot-path behavior and Go does not support same-package tests from a separate folder.

## Internal Assets

Embedded OpenAI-compatible vocab files live under `internal/openai/data`, with the embed owner in `internal/openai`.

## Optional Adapters

Adapters are separate modules under `adapters/` so root users do not pull provider-specific dependencies.

## Cacheflow

Prompt-cache boundary planning and trace simulation live under `cacheflow/`. It is a subpackage so the root tokenizer API stays focused while cache simulation can grow independently without adding dependencies.

## CLI, Docs, Tools, Benchmarks

- `cmd/omni`: CLI.
- `docs/`: user-facing documentation.
- `tools/`: development and comparison helpers.
- `benchmarks/`: reproducible benchmark harnesses and curated baselines.

## Cleanup Rule

Prefer adding folders for separate modules, commands, docs, tools, or internal assets. Avoid moving root package implementation files into folders unless creating a deliberate new package boundary with tests and API review.
