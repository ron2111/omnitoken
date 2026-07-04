#!/usr/bin/env python3
"""Generate reference corpus digests with OpenAI's Python tiktoken package.

Install the reference package first:

    python -m pip install tiktoken

Then run:

    python tools/openai_parity_digest.py
"""

from __future__ import annotations

import hashlib
import struct

import tiktoken


REFERENCE_CORPUS_SIZE = 50_000


def reference_corpus_text(i: int) -> str:
    words = ["hello", "world", "token", "cache", "scanner", "BPE", "OpenAI", "gpt-4o", "JSON", "markdown", "unicode", "throughput"]
    cjk = ["こんにちは世界", "中文测试", "안녕하세요 세계", "ภาษาไทยทดสอบ", "مرحبا بالعالم"]
    emoji = ["😀", "🚀", "👩‍💻", "🔥", "✨", "🧪", "🌍", "✅"]
    code = ["func main() { return }", "if err != nil { return err }", "const value = items[index]", "SELECT * FROM users WHERE id = 123", "for i := 0; i < n; i++ { sum += i }"]
    markdown = ["# Title\n\n- item one\n- item two", "**bold** _italic_ `code`", "> quoted text\n\n```go\nfmt.Println(x)\n```", "[link](https://example.com/path?q=token)"]
    spaces = [" ", "  ", "\t", "\n", " \n", "\r\n", "   leading", "trailing   ", "middle   gap"]

    branch = i % 16
    if branch == 0:
        return f"{words[i % len(words)]} {words[(i * 7 + 3) % len(words)]} {i}"
    if branch == 1:
        return f"I'm testing {i % 1000}{(i * 7) % 1000}{(i * 13) % 1000} tokens, you're checking counts."
    if branch == 2:
        active = "true" if i % 2 == 0 else "false"
        return f'{{"id":{i},"name":"{words[i % len(words)]}","active":{active},"score":{i * 17}}}'
    if branch == 3:
        return markdown[i % len(markdown)] + f"\n\nParagraph {i} with {words[(i + 5) % len(words)]}."
    if branch == 4:
        return code[i % len(code)] + f" // case_{i}"
    if branch == 5:
        return f"{cjk[i % len(cjk)]} {words[i % len(words)]} {emoji[i % len(emoji)]} {i}"
    if branch == 6:
        return f"snake_case/path-to/file_{i}.go::FunctionName"
    if branch == 7:
        return f"HTTPServerError{i} ABCdefGHI XYZabc"
    if branch == 8:
        return f"{spaces[i % len(spaces)]}{words[i % len(words)]}{spaces[(i + 3) % len(spaces)]}{words[(i + 4) % len(words)]}"
    if branch == 9:
        return f"Numbers: {i % 1000:03d} {i * 37 % 1000000:06d} {i * 7919:09d} {i / 7.0:.2f}"
    if branch == 10:
        return f"Symbols []{{}}()<> +=-*/ % ^ & | ~ ! ? #{i}"
    if branch == 11:
        return f"URLs/email: https://example.com/{words[i % len(words)]}/{i}?a=b&c=d user{i}@example.com"
    if branch == 12:
        return long_reference_prompt(i, words, emoji)
    if branch == 13:
        return f"Mixed scripts {cjk[(i + 1) % len(cjk)]} {emoji[(i + 2) % len(emoji)]} {words[(i + 3) % len(words)]} caf\u00e9 e\u0301 na\u00efve"
    if branch == 14:
        return f"<|endoftext|> is ordinary text here {i} <|start|><|channel|><|end|>"
    return f"line one {i}\nline two\r\n\tindented {words[i % len(words)]} {emoji[i % len(emoji)]}"


def long_reference_prompt(i: int, words: list[str], emoji: list[str]) -> str:
    text = "System: You are a tokenizer benchmark assistant."
    for j in range(20):
        text += f"\nStep {j:02d}: preserve {words[(i + j) % len(words)]}, count {i * j + j}, emit {emoji[(i + j) % len(emoji)]} safely."
    return text


def corpus_digest(encoding_name: str) -> str:
    encoding = tiktoken.get_encoding(encoding_name)
    digest = hashlib.sha256()
    for i in range(REFERENCE_CORPUS_SIZE):
        text = reference_corpus_text(i)
        tokens = encoding.encode_ordinary(text)
        digest.update(struct.pack("<I", i))
        digest.update(encoding_name.encode("utf-8"))
        digest.update(b"\x00")
        digest.update(text.encode("utf-8"))
        digest.update(b"\x00")
        digest.update(struct.pack("<I", len(tokens)))
        for token in tokens:
            digest.update(struct.pack("<I", token))
    return digest.hexdigest()


def main() -> None:
    for encoding in ("cl100k_base", "o200k_base"):
        print(f"{encoding}: {corpus_digest(encoding)}")


if __name__ == "__main__":
    main()
