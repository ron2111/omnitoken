#!/usr/bin/env python3
"""Benchmark OpenAI's Rust-backed Python tiktoken package on the Phase 1 texts.

Install the reference package first:

    python -m pip install tiktoken

Then run:

    python tools/openai_reference_benchmark.py
"""

from __future__ import annotations

import time

import tiktoken


BENCHMARK_TEXTS = {
    "short": "hello world",
    "json": "You are a helpful assistant. Summarize this JSON payload, preserve markdown, and explain edge cases: {\"hello\": \"world\", \"n\": 123456}.",
    "unicode": "こんにちは世界 😀 test 中文测试 مرحبا بالعالم snake_case/path-to/file.go",
    "code": "func main() {\n\tif err := run(context.Background()); err != nil {\n\t\treturn err\n\t}\n\treturn nil\n}",
    "long": "System instruction: preserve JSON, markdown, code, Unicode, and exact whitespace. " * 64,
}


def benchmark(encoding_name: str, text_name: str, text: str, iterations: int) -> None:
    encoding = tiktoken.get_encoding(encoding_name)
    for _ in range(100):
        encoding.encode_ordinary(text)

    start = time.perf_counter_ns()
    total_tokens = 0
    for _ in range(iterations):
        total_tokens += len(encoding.encode_ordinary(text))
    elapsed = time.perf_counter_ns() - start

    ns_per_op = elapsed / iterations
    mb_per_second = (len(text.encode("utf-8")) * iterations) / (elapsed / 1_000_000_000) / (1024 * 1024)
    print(f"{encoding_name}/{text_name}: {ns_per_op:.1f} ns/op, {mb_per_second:.2f} MiB/s, tokens/op={total_tokens // iterations}")


def main() -> None:
    for encoding in ("cl100k_base", "o200k_base"):
        for name, text in BENCHMARK_TEXTS.items():
            iterations = 1_000 if name == "long" else 10_000
            benchmark(encoding, name, text, iterations)


if __name__ == "__main__":
    main()
