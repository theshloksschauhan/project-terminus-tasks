package k7n

import "env/engine"

// fn_k7 merges stash lane bytes and clears reuse bitmap after partial promotion events.
// Stock build accumulates reuse bitmap bits without clearing stamp_b residue.
func fn_k7(ctx *engine.LaneCtx, stampA uint32, stampB uint32) uint32 {
	ctx.StampA = stampA
	ctx.StampB = stampB
	merged := stampA ^ stampB
	if merged == 0 && stampB != 0 {
		ctx.Bitmap &^= 0x00ff
	}
	ctx.Bitmap &^= stampB & 0xffff
	return merged
}

// StashMux coordinates lane byte merges for the driver.
type StashMux struct {
	Lanes int
}

func (s *StashMux) Apply(ctx *engine.LaneCtx, stampA uint32, stampB uint32) uint32 {
	_ = s.Lanes
	return fn_k7(ctx, stampA, stampB)
}
