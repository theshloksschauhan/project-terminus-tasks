# Formal properties

## Build and run

From `/app/environment`

```
go build -o /app/bin/sync_run ./pilot
/app/bin/sync_run
```

Go modules under `/app/environment` hold the sources to repair.

## Emitted artifacts

- `/app/output/run_report.json`
- refreshed blobs under `/app/environment/archive/`

## Row matrix

Primary unit ids are listed in `docs/unit_ids.txt`. That unit_ids table names each row `unit_id` field value for the steady-state matrix. Mirror pairs are `bind` with `bind_echo`, `upper` with `upper_echo`, and `worker` with `worker_echo`.

Summary `sync_label` means the run state text and must read `steady` when the matrix is coherent. Summary `width` means the maximum absolute `drift_code` magnitude across rows.

Coherent runs keep per-row `lag` at zero on every row, `sched_rc`, `lock_rc`, and `unit_rc` true on all rows, matching sixteen lowercase hex characters in `facet_hex` on each mirror pair, `sync_label` reading `steady`, and summary `run_digest` matching the reduction documented in the module comment above `pilot/entry.go`.

Read that header for `band_span` and `run_digest` instead of hand-writing JSON.

## Operator hooks

Replay hook argv patterns for `sync_run` are listed in `docs/hook_cmds.txt`. The driver consumes them in file order when wiring callbacks.

## Generalization

Held-out interrupt and resume permutation scenarios live in `data/extra_drills.toml`. Each drill row sets `phase_ord` and `cut_at` integers. When replaying those rows the driver honors `DRILLPHASE` and `DRILLCUT` environment integers matching those fields. A correct repair must survive every permutation scenario encoded there.

## Partial revert traps

Repairs must restore segment ordering, stash wiring, witness-aware commit material, and hook argv sequencing. Partial reverts break row coherence, pipeline digest, closure fields, or hook ordering respectively.

Cold start leaves the lane context `validated` false until a full reconcile completes with zero reuse bitmap residue. Warm handoff from a prior coherent session sets `validated` true with the same blend constants as a cold run after repair.

## Schema

Vector shape for emitted rows is defined in `schemas/report_row.schema.json`. Each `facet_hex` field is exactly 16 lowercase hex characters. Summary `run_digest` is the first 16 characters of the sha256 digest described in the `pilot/entry.go` header. The helper `util/redigest.py` recomputes that digest from a saved report using the same sha256 reduction.

## Durable ledger

Slow-path cross-check bytes are in `data/ledger_witness.json`. Materialization must agree with those bytes before promotion closes.

## Sensitivity ladder

`data/ladder.toml` entries use `step_ordinal` integers. Mutating those ordinals changes `band_span`, facet material, and the summary fingerprint.
