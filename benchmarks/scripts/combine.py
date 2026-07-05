#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import json


FIELDS = ["runner", "operation", "encoding", "case", "ns_per_op", "mb_per_s", "b_per_op", "allocs_per_op", "source"]


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--inputs", required=True, help="Comma-separated JSONL/CSV inputs")
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    rows = []
    for path in args.inputs.split(","):
        path = path.strip()
        if not path:
            continue
        if path.endswith(".csv"):
            with open(path, "r", encoding="utf-8", newline="") as f:
                rows.extend(csv.DictReader(f))
        else:
            with open(path, "r", encoding="utf-8") as f:
                for line in f:
                    if line.strip():
                        rows.append(json.loads(line))

    with open(args.output, "w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=FIELDS)
        writer.writeheader()
        for row in rows:
            writer.writerow({field: row.get(field, "") for field in FIELDS})


if __name__ == "__main__":
    main()
