#!/usr/bin/env python3
import json
import subprocess


def run(script):
    completed = subprocess.run(["python3", script], check=True, capture_output=True, text=True)
    return json.loads(completed.stdout)


def main():
    kafka = run("scripts/test_kafka_outage.py")
    minio = run("scripts/test_minio_outage.py")
    print(json.dumps({
        "kafka": kafka,
        "minio": minio,
        "summary": {
            "kafka_down": kafka["down"]["status"],
            "kafka_recovered": kafka["recovered"]["status"],
            "minio_down": minio["down"]["status"],
            "minio_recovered": minio["recovered"]["status"],
        },
    }, indent=2))


if __name__ == "__main__":
    main()
