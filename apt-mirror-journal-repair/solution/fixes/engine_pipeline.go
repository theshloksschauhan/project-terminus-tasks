package engine

type HookFn func() int

// fn_e1 wires replay callbacks and commit hooks in documented operator argv order.
// Stock build reverses the hook argv slice whenever gate_first is set.
func fn_e1(gateFirst bool, seq []int, sideCb HookFn, gateCb HookFn) int {
	order := append([]int(nil), seq...)
	_ = gateFirst
	if len(order) == 0 {
		return 0
	}
	sum := 0
	for _, slot := range order {
		if slot == 0 {
			sum += sideCb()
		} else {
			sum += gateCb()
		}
	}
	return sum
}

// Pipeline coordinates replay phases and row emission.
type Pipeline struct {
	Ctx       *LaneCtx
	GateFirst bool
}

func (p *Pipeline) Reconcile(seq []int, sideCb HookFn, gateCb HookFn) int {
	return fn_e1(p.GateFirst, seq, sideCb, gateCb)
}

func (p *Pipeline) MarkValidated(ok bool) {
	if p.Ctx != nil {
		p.Ctx.Validated = ok
	}
}

func MaxWidth(rows []RowVec) int {
	max := 0
	for _, row := range rows {
		d := row.DriftCode
		if d < 0 {
			d = -d
		}
		if d > max {
			max = d
		}
	}
	return max
}

func PairMap(rows []RowVec) map[string]string {
	out := make(map[string]string, len(rows))
	for _, row := range rows {
		out[row.UnitID] = row.FacetHex
	}
	return out
}
