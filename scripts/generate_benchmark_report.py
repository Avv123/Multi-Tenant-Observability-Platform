#!/usr/bin/env python3
import json
import pathlib
import sys
from datetime import datetime, timezone


def load(path):
    with open(path, "r", encoding="utf-8") as handle:
        return json.load(handle)


def main():
    if len(sys.argv) < 3:
        raise SystemExit("usage: generate_benchmark_report.py <output-dir> <json-file> [<json-file>...]")
    output_dir = pathlib.Path(sys.argv[1])
    output_dir.mkdir(parents=True, exist_ok=True)
    reports = {pathlib.Path(path).stem: load(path) for path in sys.argv[2:]}
    timestamp = datetime.now(timezone.utc).strftime("%Y%m%dT%H%M%SZ")

    summary = {
        "generated_at": timestamp,
        "reports": reports,
    }
    (output_dir / f"benchmark-summary-{timestamp}.json").write_text(json.dumps(summary, indent=2))

    lines = ["# PulseLens Local Benchmark Summary", "", f"Generated at: `{timestamp}`", ""]
    for name, report in reports.items():
        lines.append(f"## {name}")
        for key, value in report.items():
            lines.append(f"- `{key}`: `{value}`")
        lines.append("")
    (output_dir / f"benchmark-summary-{timestamp}.md").write_text("\n".join(lines))
    print(json.dumps({"timestamp": timestamp, "output_dir": str(output_dir)}, indent=2))


if __name__ == "__main__":
    main()
