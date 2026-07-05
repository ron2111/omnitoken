#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import json
import os
from collections import defaultdict


COLORS = ["#2563eb", "#dc2626", "#16a34a", "#9333ea", "#ea580c", "#0891b2"]


def read_rows(path: str) -> list[dict[str, str]]:
    with open(path, "r", encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f))


def svg_bar(rows: list[dict[str, str]], operation: str, metric: str, out_path: str) -> None:
    selected = [r for r in rows if r["operation"] == operation and r.get(metric)]
    labels = [f"{r['runner']}\n{r['encoding']}\n{r['case']}" for r in selected]
    values = [float(r[metric]) for r in selected]
    if not values:
        return
    width = max(900, len(values) * 72)
    height = 420
    margin = 60
    maxv = max(values) or 1
    barw = max(18, (width - 2 * margin) / max(1, len(values)) * 0.7)
    step = (width - 2 * margin) / max(1, len(values))
    parts = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">']
    parts.append('<rect width="100%" height="100%" fill="white"/>')
    parts.append(f'<text x="{margin}" y="30" font-family="sans-serif" font-size="20" font-weight="700">{operation} {metric}</text>')
    parts.append(f'<line x1="{margin}" y1="{height-margin}" x2="{width-margin}" y2="{height-margin}" stroke="#111827"/>')
    for i, (label, value) in enumerate(zip(labels, values)):
        x = margin + i * step + (step - barw) / 2
        h = (height - 2 * margin - 40) * value / maxv
        y = height - margin - h
        color = COLORS[i % len(COLORS)]
        parts.append(f'<rect x="{x:.1f}" y="{y:.1f}" width="{barw:.1f}" height="{h:.1f}" fill="{color}"/>')
        parts.append(f'<text x="{x + barw/2:.1f}" y="{y - 4:.1f}" text-anchor="middle" font-family="sans-serif" font-size="10">{value:.0f}</text>')
        short = label.replace("\n", " / ")[:28]
        parts.append(f'<text x="{x + barw/2:.1f}" y="{height - margin + 14}" text-anchor="middle" font-family="sans-serif" font-size="9" transform="rotate(45 {x + barw/2:.1f},{height - margin + 14})">{short}</text>')
    parts.append("</svg>")
    with open(out_path, "w", encoding="utf-8") as f:
        f.write("\n".join(parts))


def summary(rows: list[dict[str, str]], out_path: str, metadata_path: str | None = None) -> None:
    by_op: dict[str, list[dict[str, str]]] = defaultdict(list)
    for row in rows:
        by_op[row["operation"]].append(row)
    lines = ["# Benchmark Summary", "", "Generated from `benchmarks/results/combined.csv`.", ""]
    if metadata_path and os.path.exists(metadata_path):
        with open(metadata_path, "r", encoding="utf-8-sig") as f:
            meta = json.load(f)
        lines.extend([
            "## Run Metadata",
            "",
            "| Field | Value |",
            "| --- | --- |",
            f"| Timestamp UTC | `{meta.get('timestamp_utc', '')}` |",
            f"| Git commit | `{meta.get('git_commit', '')}` |",
            f"| Go | `{meta.get('go_version', '')}` |",
            f"| OS | `{meta.get('os', {}).get('caption', '')} {meta.get('os', {}).get('version', '')}` |",
            f"| CPU | `{meta.get('cpu', {}).get('name', '')}` |",
            f"| CPU cores / logical | `{meta.get('cpu', {}).get('cores', '')} / {meta.get('cpu', {}).get('logical_processors', '')}` |",
            f"| CPU load sample | `{meta.get('cpu', {}).get('load_percent_sample', '')}` |",
            f"| Total memory bytes | `{meta.get('memory', {}).get('total_physical_bytes', '')}` |",
            f"| Free memory KB | `{meta.get('memory', {}).get('free_physical_kb', '')}` |",
            f"| Docker | `{meta.get('docker_version', '')}` |",
            f"| Benchmark settings | `count={meta.get('benchmark', {}).get('count', '')}, benchtime={meta.get('benchmark', {}).get('benchtime', '')}, rust={meta.get('benchmark', {}).get('rust', '')}` |",
            "",
            f"> {meta.get('benchmark', {}).get('note', '')}",
            "",
        ])
    for op, op_rows in sorted(by_op.items()):
        lines.extend([f"## {op}", "", "| Runner | Encoding | Case | ns/op | MB/s | B/op | allocs/op |", "| --- | --- | --- | ---: | ---: | ---: | ---: |"])
        for r in sorted(op_rows, key=lambda x: (x["encoding"], x["case"], x["runner"])):
            lines.append(f"| {r['runner']} | {r['encoding']} | {r['case']} | {r['ns_per_op']} | {r['mb_per_s']} | {r['b_per_op']} | {r['allocs_per_op']} |")
        lines.append("")
    with open(out_path, "w", encoding="utf-8") as f:
        f.write("\n".join(lines))


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True)
    parser.add_argument("--metadata", default="")
    parser.add_argument("--out-dir", required=True)
    args = parser.parse_args()
    os.makedirs(args.out_dir, exist_ok=True)
    rows = read_rows(args.input)
    summary(rows, os.path.join(args.out_dir, "summary.md"), args.metadata or None)
    for op in sorted({r["operation"] for r in rows}):
        svg_bar(rows, op, "ns_per_op", os.path.join(args.out_dir, f"{op}_ns_per_op.svg"))
        svg_bar(rows, op, "mb_per_s", os.path.join(args.out_dir, f"{op}_mb_per_s.svg"))


if __name__ == "__main__":
    main()
