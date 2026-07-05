#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import re


BENCH_RE = re.compile(
    r"^(Benchmark\S+)\s+\d+\s+([0-9.]+)\s+ns/op(?:\s+([0-9.]+)\s+MB/s)?(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?"
)


def parse_name(name: str) -> dict[str, str]:
    parts = name.split("/")
    out: dict[str, str] = {}
    for part in parts[1:]:
        part = part.split("-")[0]
        if "=" in part:
            k, v = part.split("=", 1)
            out[k] = v
    return out


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--runner-suffix", default="")
    args = parser.parse_args()

    rows = []
    with open(args.input, "r", encoding="utf-8") as f:
        for line in f:
            m = BENCH_RE.match(line.strip())
            if not m:
                continue
            meta = parse_name(m.group(1))
            rows.append(
                {
                    "runner": meta.get("runner", "unknown") + args.runner_suffix,
                    "operation": meta.get("op", "unknown"),
                    "encoding": meta.get("enc", "unknown"),
                    "case": meta.get("case", "unknown"),
                    "ns_per_op": float(m.group(2)),
                    "mb_per_s": float(m.group(3) or 0),
                    "b_per_op": int(m.group(4) or 0),
                    "allocs_per_op": int(m.group(5) or 0),
                    "source": "go test",
                }
            )

    with open(args.output, "w", encoding="utf-8") as f:
        for row in rows:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")


if __name__ == "__main__":
    main()
