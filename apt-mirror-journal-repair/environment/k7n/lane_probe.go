package k7n

import "env/engine"

// LaneProbe emits diagnostic lane probes without stash authority.
type LaneProbe struct {
	Tag string
}

func (p *LaneProbe) Sample(ctx *engine.LaneCtx) int {
	return int(ctx.StampA&0xff) ^ int(ctx.StampB&0xff)
}

func (p *LaneProbe) String() string {
	return p.Tag
}
