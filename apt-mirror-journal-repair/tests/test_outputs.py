"""Domain verifier for apt-mirror-journal-repair."""

from __future__ import annotations

import json
import os
import re
import shutil
import subprocess
from pathlib import Path

ENV = Path("/app/environment")
OUT = Path("/app/output/run_report.json")
REDIGEST = ENV / "util/redigest.py"
UNIT_IDS = ENV / "docs/unit_ids.txt"
LADDER = ENV / "data/ladder.toml"
EXTRA = ENV / "data/extra_drills.toml"
HEX16 = re.compile(r"^[0-9a-f]{16}$")


def _run_sync(extra_env: dict[str, str] | None = None) -> None:
    env = os.environ.copy()
    if extra_env:
        env.update(extra_env)
    subprocess.run(
        ["go", "build", "-o", "/app/bin/sync_run", "./pilot"],
        cwd=ENV,
        check=True,
        env=env,
    )
    subprocess.run(["/app/bin/sync_run"], check=True, env=env)


def run_pipeline(extra_env: dict[str, str] | None = None) -> None:
    _run_sync(extra_env)


def load_report() -> dict:
    return json.loads(OUT.read_text())


def digest_from_report(report: dict | None = None) -> str:
    path = OUT
    if report is not None:
        path = OUT.with_suffix(".tmp.json")
        path.write_text(json.dumps(report))
    try:
        proc = subprocess.run(
            ["python3", str(REDIGEST), str(path), str(ENV)],
            check=True,
            capture_output=True,
            text=True,
        )
        return proc.stdout.strip()
    finally:
        if report is not None:
            path.unlink(missing_ok=True)


def read_unit_ids() -> list[str]:
    lines = []
    for line in UNIT_IDS.read_text().splitlines():
        line = line.strip()
        if line and not line.startswith("#"):
            lines.append(line)
    return lines


def coherent_rows(rows: list[dict]) -> bool:
    pairs = {
        "bind": "bind_echo",
        "upper": "upper_echo",
        "worker": "worker_echo",
    }
    by_id = {row["unit_id"]: row for row in rows}
    for left, right in pairs.items():
        if by_id[left]["facet_hex"] != by_id[right]["facet_hex"]:
            return False
    for row in rows:
        if row["lag"] != 0:
            return False
        if not (row["sched_rc"] and row["lock_rc"] and row["unit_rc"]):
            return False
        if not HEX16.match(row["facet_hex"]):
            return False
    return True


def parse_extra_drills() -> list[dict[str, int | str]]:
    drills: list[dict[str, int | str]] = []
    block: dict[str, int | str] = {}
    for line in EXTRA.read_text().splitlines():
        line = line.strip()
        if line == "[[drill]]":
            if block:
                drills.append(block)
            block = {}
            continue
        m = re.match(r"(\w+)\s*=\s*(\d+)", line)
        if m:
            block[m.group(1)] = int(m.group(2))
        m2 = re.match(r'name\s*=\s*"([^"]+)"', line)
        if m2:
            block["name"] = m2.group(1)
    if block:
        drills.append(block)
    return drills



def test_p1_shape_vec() -> None:
    """Row vector length and unit_id ordering match unit_ids.txt."""
    run_pipeline()
    report = load_report()
    ids = read_unit_ids()
    got = [row["unit_id"] for row in report["rows"]]
    assert got == ids


def test_p2_hex_lower() -> None:
    """facet_hex material uses sixteen lowercase hex characters."""
    run_pipeline()
    report = load_report()
    for row in report["rows"]:
        assert HEX16.match(row["facet_hex"])


def test_p3_rc_fields() -> None:
    """Coherent emission keeps lag zero and rc fields true on every row."""
    run_pipeline()
    report = load_report()
    assert coherent_rows(report["rows"])


def test_p4_label_text() -> None:
    """Summary sync_label reads steady when the matrix is coherent."""
    run_pipeline()
    report = load_report()
    assert report["summary"]["sync_label"] == "steady"


def test_p5_width_max() -> None:
    """Summary width equals the maximum absolute drift_code magnitude."""
    run_pipeline()
    report = load_report()
    width = max(abs(row["drift_code"]) for row in report["rows"])
    assert report["summary"]["width"] == width


def test_p6_reduce_hex() -> None:
    """Summary run_digest matches the pilot header reduction from rows."""
    run_pipeline()
    report = load_report()
    assert report["summary"]["run_digest"] == digest_from_report(report)


def test_p7_known_good() -> None:
    """Repaired pipeline fingerprint matches the public digest reduction."""
    run_pipeline()
    report = load_report()
    assert coherent_rows(report["rows"])
    assert report["summary"]["run_digest"] == digest_from_report(report)


def test_p8_dual_run() -> None:
    """Consecutive pipeline runs emit identical fingerprints."""
    run_pipeline()
    first = load_report()["summary"]["run_digest"]
    run_pipeline()
    second = load_report()["summary"]["run_digest"]
    assert first == second


def test_p9_sens_steps() -> None:
    """Mutating ladder step ordinals changes facet material and fingerprint."""
    run_pipeline()
    base = load_report()["summary"]["run_digest"]
    backup = LADDER.read_text()
    try:
        LADDER.write_text(backup.replace("step_ordinal = 3", "step_ordinal = 4", 1))
        run_pipeline()
        mutated = load_report()["summary"]["run_digest"]
    finally:
        LADDER.write_text(backup)
    assert mutated != base


def test_p10_rerun_pipe() -> None:
    """Hand-written JSON is replaced when the pipeline reruns."""
    run_pipeline()
    OUT.write_text(json.dumps({"rows": [], "summary": {"sync_label": "steady", "run_digest": "0" * 16, "width": 1}}))
    run_pipeline()
    report = load_report()
    assert coherent_rows(report["rows"])
    assert report["summary"]["run_digest"] == digest_from_report(report)
    assert len(report["rows"]) == len(read_unit_ids())



def test_p14_all_extras() -> None:
    """Every held-out permutation scenario in extra_drills.toml survives after fix."""
    for drill in parse_extra_drills():
        phase = str(drill.get("phase_ord", 0))
        cut = str(drill.get("cut_at", 0))
        run_pipeline({"DRILLPHASE": phase, "DRILLCUT": cut})
        report = load_report()
        assert coherent_rows(report["rows"])
        assert report["summary"]["sync_label"] == "steady"
        assert report["summary"]["run_digest"] == digest_from_report(report)


def test_p15_premature() -> None:
    """Steady sync_label requires agreeing mirror pairs and coherent rows."""
    run_pipeline()
    report = load_report()
    if report["summary"]["sync_label"] == "steady":
        assert coherent_rows(report["rows"])


def test_p16_cli_hooks() -> None:
    """hook_cmds argv ordering is required for sync_run hook success."""
    hooks = ENV / "docs/hook_cmds.txt"
    backup = hooks.read_text()
    try:
        hooks.write_text("--gate-first\n--gate-cb\n--side-cb\n")
        run_pipeline()
        flipped_digest = load_report()["summary"]["run_digest"]
        hooks.write_text(backup)
        run_pipeline()
        normal_report = load_report()
        normal_digest = normal_report["summary"]["run_digest"]
        assert flipped_digest != normal_digest
        assert coherent_rows(normal_report["rows"])
        assert normal_digest == digest_from_report(normal_report)
    finally:
        hooks.write_text(backup)


def setup_module() -> None:
    shutil.rmtree(OUT.parent, ignore_errors=True)
    OUT.parent.mkdir(parents=True, exist_ok=True)
