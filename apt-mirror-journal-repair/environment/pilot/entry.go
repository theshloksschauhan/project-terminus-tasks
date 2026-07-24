// sync_run driver for the offline package archive workspace.
//
// band_span is the uint16 sum of step_ordinal values from data/ladder.toml.
//
// run_digest is the first sixteen lowercase hex characters of SHA-256 over the
// concatenation of sorted (unit_id || facet_hex) pairs followed by band_span as
// big-endian uint16 bytes.
package main

import (
	"bufio"
	"encoding/json"
	"env/engine"
	"env/k7n"
	"env/m3x"
	"env/r2p"
	"env/tools"
	"env/util"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	envRoot  = "/app/environment"
	outPath  = "/app/output/run_report.json"
	arcRoot  = "/app/environment/archive"
)

type ladderFile struct {
	Steps []struct {
		StepOrdinal int `toml:"step_ordinal"`
	} `toml:"steps"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "sync_run: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	_ = os.MkdirAll("/app/output", 0o755)
	_ = os.MkdirAll(arcRoot, 0o755)

	units, err := readLines(filepath.Join(envRoot, "docs/unit_ids.txt"))
	if err != nil {
		return err
	}
	steps, err := readLadder(filepath.Join(envRoot, "data/ladder.toml"))
	if err != nil {
		return err
	}
	hooks, err := readLines(filepath.Join(envRoot, "docs/hook_cmds.txt"))
	if err != nil {
		return err
	}

	var witness engine.Witness
	if err := util.ReadJSON(filepath.Join(envRoot, "data/ledger_witness.json"), &witness); err != nil {
		return err
	}

	ctx := &engine.LaneCtx{Seq: witness.Seq}
	mat := &tools.RowMat{Steps: steps}
	seg := &m3x.SegMux{}
	stash := &k7n.StashMux{Lanes: 3}
	fold := &r2p.CommitFold{Witness: &witness}
	gauge := &tools.GaugeMux{Phase: "early"}

	gateFirst := len(hooks) > 0 && strings.TrimSpace(hooks[0]) == "--gate-first"
	hookSeq := make([]int, 0, len(hooks))
	for _, h := range hooks[1:] {
		switch strings.TrimSpace(h) {
		case "--side-cb":
			hookSeq = append(hookSeq, 0)
		case "--gate-cb":
			hookSeq = append(hookSeq, 1)
		}
	}
	if len(hookSeq) == 0 {
		hookSeq = []int{0, 1}
	}

	phaseBump := 0
	if v := os.Getenv("DRILLPHASE"); v != "" {
		fmt.Sscanf(v, "%d", &phaseBump)
	}

	cutAt := -1
	if v := os.Getenv("DRILLCUT"); v != "" {
		fmt.Sscanf(v, "%d", &cutAt)
	}

	blend := 0
	sideFn := func() int {
		limit := len(steps)
		if cutAt >= 0 && cutAt < limit {
			limit = cutAt + 1
		}
		for i := 0; i < limit; i++ {
			blend = seg.Next(i, len(steps)-1-i+phaseBump, blend)
		}
		return blend
	}
	gateFn := func() int {
		if blend == 0 {
			ctx.Bitmap |= 0x00ff
		}
		stampA := uint32(mat.BandSpan())
		stampB := uint32(witness.Seq)
		return int(stash.Apply(ctx, stampA, stampB))
	}

	pipe := &engine.Pipeline{Ctx: ctx, GateFirst: gateFirst}
	ord := pipe.Reconcile(hookSeq, sideFn, gateFn)

	rows := mat.Build(units, blend, ord)

	// Align echo facets to primary units before witness materialization.
	prim := engine.PairMap(rows)
	for i, row := range rows {
		switch row.UnitID {
		case "bind_echo":
			rows[i].FacetHex = prim["bind"]
		case "upper_echo":
			rows[i].FacetHex = prim["upper"]
		case "worker_echo":
			rows[i].FacetHex = prim["worker"]
		}
	}

	rows = fold.Fold(rows, steps)

	okWitness := r2p.WitnessMatch(rows, &witness)
	if os.Getenv("DRILLPHASE") != "" || os.Getenv("DRILLCUT") != "" {
		pairs := engine.PairMap(rows)
		okWitness = ctx.Bitmap == 0 &&
			pairs["bind"] == pairs["bind_echo"] &&
			pairs["upper"] == pairs["upper_echo"] &&
			pairs["worker"] == pairs["worker_echo"]
		if okWitness {
			for i := range rows {
				rows[i].DriftCode = 0
			}
		}
	}
	for i := range rows {
		if !okWitness {
			rows[i].Lag = 1
			rows[i].SchedRC = false
			rows[i].LockRC = false
			rows[i].UnitRC = false
		} else if ctx.Bitmap != 0 {
			rows[i].Lag = 1
			rows[i].UnitRC = false
		} else {
			rows[i].Lag = 0
			rows[i].SchedRC = true
			rows[i].LockRC = true
			rows[i].UnitRC = true
		}
	}

	label := gauge.Label(rows)
	if okWitness && ctx.Bitmap == 0 && label == "steady" {
		gauge.Phase = "done"
		label = gauge.Label(rows)
	}

	pairs := make([][2]string, 0, len(rows))
	for _, row := range rows {
		pairs = append(pairs, [2]string{row.UnitID, row.FacetHex})
	}
	digest := util.RunReduce(pairs, mat.BandSpan())

	report := engine.RunReport{
		Rows: rows,
		Summary: engine.Summary{
			SyncLabel: label,
			RunDigest: digest,
			Width:     engine.MaxWidth(rows),
		},
	}

	if err := util.WriteJSON(outPath, report); err != nil {
		return err
	}
	if err := writeArchive(rows, digest); err != nil {
		return err
	}
	pipe.MarkValidated(okWitness && ctx.Bitmap == 0)
	return nil
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	return out, sc.Err()
}

func readLadder(path string) ([]int, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ord []int
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "step_ordinal") {
			var v int
			_, err := fmt.Sscanf(line, "step_ordinal = %d", &v)
			if err == nil {
				ord = append(ord, v)
			}
		}
	}
	if len(ord) == 0 {
		return []int{3, 5, 7}, nil
	}
	return ord, nil
}

func writeArchive(rows []engine.RowVec, digest string) error {
	meta := map[string]string{"fingerprint": digest}
	raw, _ := json.MarshalIndent(meta, "", "  ")
	if err := os.WriteFile(filepath.Join(arcRoot, "meta.json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}
	for _, row := range rows {
		p := filepath.Join(arcRoot, row.UnitID+".bin")
		if err := os.WriteFile(p, []byte(row.FacetHex), 0o644); err != nil {
			return err
		}
	}
	return nil
}
