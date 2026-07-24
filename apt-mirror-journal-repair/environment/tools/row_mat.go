package tools

import (
	"env/engine"
	"env/util"
)

// RowMat builds row vectors from unit ids and lane blend state.
type RowMat struct {
	Steps []int
}

func (r *RowMat) Build(units []string, blend int, ord int) []engine.RowVec {
	rows := make([]engine.RowVec, 0, len(units))
	for _, unit := range units {
		facet := util.FacetMaterial(unit, blend, ord)
		rows = append(rows, engine.RowVec{
			UnitID:    unit,
			Lag:       0,
			SchedRC:   true,
			LockRC:    true,
			UnitRC:    true,
			FacetHex:  facet,
			DriftCode: 0,
		})
	}
	return rows
}

func (r *RowMat) BandSpan() uint16 {
	var span uint16
	for _, step := range r.Steps {
		span += uint16(step)
	}
	return span
}
