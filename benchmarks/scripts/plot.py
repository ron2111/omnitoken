#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import json
import os
from collections import defaultdict
from math import exp, log


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


def geometric_mean(values: list[float]) -> float:
    values = [v for v in values if v > 0]
    if not values:
        return 0
    return exp(sum(log(v) for v in values) / len(values))


def speedup(rows: list[dict[str, str]], left_runner: str, left_op: str, right_runner: str, right_op: str) -> tuple[float, int]:
    left = {
        (r["encoding"], r["case"]): float(r["ns_per_op"])
        for r in rows
        if r["runner"] == left_runner and r["operation"] == left_op and r.get("ns_per_op")
    }
    right = {
        (r["encoding"], r["case"]): float(r["ns_per_op"])
        for r in rows
        if r["runner"] == right_runner and r["operation"] == right_op and r.get("ns_per_op")
    }
    ratios = [right[key] / left_value for key, left_value in left.items() if key in right and left_value > 0]
    return geometric_mean(ratios), len(ratios)


def speedup_summary(rows: list[dict[str, str]]) -> list[tuple[str, float, int]]:
    comparisons = [
        ("Native count vs tiktoken-go", "omnitoken", "count", "tiktoken_go", "count_by_encode"),
        ("Native encode vs tiktoken-go", "omnitoken", "encode", "tiktoken_go", "encode"),
        ("Native decode vs tiktoken-go", "omnitoken", "decode", "tiktoken_go", "decode"),
        ("Docker count vs Docker tiktoken-go", "omnitoken_docker", "count", "tiktoken_go_docker", "count_by_encode"),
        ("Docker encode vs Docker tiktoken-go", "omnitoken_docker", "encode", "tiktoken_go_docker", "encode"),
        ("Docker decode vs Docker tiktoken-go", "omnitoken_docker", "decode", "tiktoken_go_docker", "decode"),
        ("Docker encode vs OpenAI Rust", "omnitoken_docker", "encode", "openai_rust", "encode"),
        ("Docker count vs OpenAI Rust count-by-encode", "omnitoken_docker", "count", "openai_rust", "count_by_encode"),
    ]
    out = []
    for label, left_runner, left_op, right_runner, right_op in comparisons:
        ratio, n = speedup(rows, left_runner, left_op, right_runner, right_op)
        if n:
            out.append((label, ratio, n))
    return out


def best_marketing_speedups(rows: list[dict[str, str]]) -> list[tuple[str, float, int]]:
    candidates = [
        ("vs tiktoken-go / Count", "omnitoken", "count", "tiktoken_go", "count_by_encode"),
        ("vs tiktoken-go / Encode", "omnitoken", "encode", "tiktoken_go", "encode"),
        ("vs tiktoken-go / Decode", "omnitoken", "decode", "tiktoken_go", "decode"),
        ("vs OpenAI Rust / Count", "omnitoken_docker", "count", "openai_rust", "count_by_encode"),
        ("vs OpenAI Rust / Encode", "omnitoken_docker", "encode", "openai_rust", "encode"),
    ]
    out = []
    for label, left_runner, left_op, right_runner, right_op in candidates:
        ratio, n = speedup(rows, left_runner, left_op, right_runner, right_op)
        if n:
            out.append((label, ratio, n))
    return out


def svg_speedups(speedups: list[tuple[str, float, int]], out_path: str) -> None:
    if not speedups:
        return
    width = 1000
    row_h = 44
    margin_l = 300
    margin_r = 60
    height = 70 + row_h * len(speedups)
    maxv = max(v for _, v, _ in speedups) or 1
    parts = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">']
    parts.append('<rect width="100%" height="100%" fill="white"/>')
    parts.append('<text x="24" y="34" font-family="sans-serif" font-size="22" font-weight="700">OmniToken geomean speedups</text>')
    parts.append('<text x="24" y="56" font-family="sans-serif" font-size="12" fill="#4b5563">Higher is better. Ratios compare ns/op across matching encoding/corpus cases.</text>')
    bar_max = width - margin_l - margin_r
    for i, (label, value, n) in enumerate(speedups):
        y = 88 + i * row_h
        bar_w = bar_max * value / maxv
        color = COLORS[i % len(COLORS)]
        parts.append(f'<text x="24" y="{y + 17}" font-family="sans-serif" font-size="13">{label}</text>')
        parts.append(f'<rect x="{margin_l}" y="{y}" width="{bar_w:.1f}" height="24" fill="{color}" rx="3"/>')
        parts.append(f'<text x="{margin_l + bar_w + 8:.1f}" y="{y + 17}" font-family="sans-serif" font-size="13" font-weight="700">{value:.2f}x</text>')
        parts.append(f'<text x="{width - margin_r}" y="{y + 17}" text-anchor="end" font-family="sans-serif" font-size="11" fill="#6b7280">n={n}</text>')
    parts.append("</svg>")
    with open(out_path, "w", encoding="utf-8") as f:
        f.write("\n".join(parts))


def svg_marketing_speedups(speedups: list[tuple[str, float, int]], out_path: str) -> None:
    if not speedups:
        return
    width = 920
    height = 360
    margin_l = 210
    margin_r = 90
    top = 86
    row_h = 46
    maxv = max(v for _, v, _ in speedups) or 1
    parts = [f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">']
    parts.append('<rect width="100%" height="100%" fill="#0b1020"/>')
    parts.append('<text x="30" y="38" font-family="Inter,Segoe UI,Arial,sans-serif" font-size="24" fill="white" font-weight="800">OmniToken benchmark speedups</text>')
    parts.append('<text x="30" y="62" font-family="Inter,Segoe UI,Arial,sans-serif" font-size="13" fill="#a7b0c0">Geomean across matching OpenAI BPE corpus cases. Higher is better.</text>')
    bar_max = width - margin_l - margin_r
    for i, (label, value, n) in enumerate(speedups):
        y = top + i * row_h
        bar_w = bar_max * value / maxv
        color = "#38bdf8" if "Rust" in label else "#22c55e"
        parts.append(f'<text x="30" y="{y + 19}" font-family="Inter,Segoe UI,Arial,sans-serif" font-size="14" fill="#e5e7eb">{label}</text>')
        parts.append(f'<rect x="{margin_l}" y="{y}" width="{bar_w:.1f}" height="26" fill="{color}" rx="6"/>')
        parts.append(f'<text x="{margin_l + bar_w + 10:.1f}" y="{y + 19}" font-family="Inter,Segoe UI,Arial,sans-serif" font-size="15" fill="white" font-weight="800">{value:.2f}x</text>')
        parts.append(f'<text x="{width - 28}" y="{y + 19}" text-anchor="end" font-family="Inter,Segoe UI,Arial,sans-serif" font-size="11" fill="#94a3b8">n={n}</text>')
    parts.append('<text x="30" y="335" font-family="Inter,Segoe UI,Arial,sans-serif" font-size="11" fill="#64748b">Run benchmarks locally with benchmarks/scripts/run.ps1 for machine-specific results.</text>')
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
    speedups = speedup_summary(rows)
    if speedups:
        lines.extend(["## Speedup Highlights", "", "| Comparison | Geomean speedup | Matched cases |", "| --- | ---: | ---: |"])
        for label, ratio, n in speedups:
            lines.append(f"| {label} | {ratio:.2f}x | {n} |")
        lines.append("")
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
    svg_speedups(speedup_summary(rows), os.path.join(args.out_dir, "speedups.svg"))
    svg_marketing_speedups(best_marketing_speedups(rows), os.path.join(args.out_dir, "speedups-readme.svg"))
    for op in sorted({r["operation"] for r in rows}):
        svg_bar(rows, op, "ns_per_op", os.path.join(args.out_dir, f"{op}_ns_per_op.svg"))
        svg_bar(rows, op, "mb_per_s", os.path.join(args.out_dir, f"{op}_mb_per_s.svg"))


if __name__ == "__main__":
    main()
