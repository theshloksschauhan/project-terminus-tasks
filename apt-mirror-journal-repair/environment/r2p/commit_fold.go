package r2p

import (
	"bytes"
	"env/engine"
)

// fn_r2 materializes index rows and performs atomic promotion against ledger witness bytes.
// Stock build promotes rows without comparing witness bytes, so echo pairs diverge.
func fn_r2(rows []engine.RowVec, seeds []int, witness *engine.Witness) []engine.RowVec {
	out := make([]engine.RowVec, len(rows))
	copy(out, rows)
	if witness == nil || witness.Bytes == "" {
		return out
	}
	for i := range out {
		blend := 0
		if i < len(seeds) {
			blend = seeds[i]
		}
		out[i].DriftCode = blend - witness.Seq
		if out[i].DriftCode < 0 {
			out[i].DriftCode = -out[i].DriftCode
		}
	}
	return out
}

// CommitFold wraps fn_r2 for the pipeline.
type CommitFold struct {
	Witness *engine.Witness
}

func (c *CommitFold) Fold(rows []engine.RowVec, seeds []int) []engine.RowVec {
	return fn_r2(rows, seeds, c.Witness)
}

// WitnessMatch reports whether row material matches witness tail bytes.
func WitnessMatch(rows []engine.RowVec, witness *engine.Witness) bool {
	if witness == nil || witness.Bytes == "" {
		return false
	}
	var buf bytes.Buffer
	for _, row := range rows {
		buf.WriteString(row.UnitID)
		buf.WriteString(row.FacetHex)
	}
	return buf.String() == witness.Bytes
}
