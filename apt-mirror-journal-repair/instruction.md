The offline package archive sync daemon requires configuration to correctly manage state when long pulls are interrupted. Currently, the synchronization process does not align partial trees with signing metadata, resulting in duplicate payload fetches, mismatched index digests, and inconsistent cold restart states.

Your goal is to configure the daemon's synchronization behavior in the `/app/environment` directory to ensure deterministic convergence across all network drills.

### Expected Output
When the daemon is correctly configured, it must produce the following artifacts:
- A `/app/output/run_report.json` report containing a `rows` vector (`unit_id`, `lag`, `sched_rc`, `lock_rc`, `unit_rc`, `facet_hex`, `drift_code`) and a `summary` block (`sync_label`, `width`, `run_digest`). 
- Refreshed tree artifacts located under `/app/environment/archive/`.

### Requirements and Constraints
- **Correct logic:** Hand-written JSON or archive blobs are insufficient; the normal application build and runtime pipeline must be capable of generating the correct artifacts from the daemon's operations. 
- **Reference documentation:** Read `/app/environment/docs/formal_rules.md` for the details concerning the row matrix, summary fields, ladder sensitivity, extra drill generalization, and the `run_digest` reduction in the module comment above `/app/environment/pilot/entry.go`. 
- **Robustness:** A correct configuration must survive every permutation scenario encoded in `/app/environment/data/extra_drills.toml`.
- **Hook order sensitivity:** The `hook_cmds.txt` argument ordering behavior is intentional and must be preserved by the system. Normal and flipped hook orders should not produce identical digests.
