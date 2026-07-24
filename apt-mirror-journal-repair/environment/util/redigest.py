
"""Recompute run_digest from a run_report.json path using the public contract."""

from __future__ import annotations

import hashlib
import json
import re
import struct
import sys
from pathlib import Path


def ladder_ordinals(ladder: Path) -> list[int]:
    vals: list[int] = []
    for line in ladder.read_text().splitlines():
        m = re.match(r"step_ordinal\s*=\s*(\d+)", line.strip())
        if m:
            vals.append(int(m.group(1)))
    return vals


def main() -> int:
    report_path = Path(sys.argv[1])
    env_root = Path(sys.argv[2]) if len(sys.argv) > 2 else Path("/app/environment")
    report = json.loads(report_path.read_text())
    span = sum(ladder_ordinals(env_root / "data/ladder.toml")) & 0xFFFF
    parts = sorted(row["unit_id"] + row["facet_hex"] for row in report["rows"])
    payload = "".join(parts).encode() + struct.pack(">H", span)
    print(hashlib.sha256(payload).hexdigest()[:16])
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
