package m3x

import "sort"

// SegmentBuffer collects ephemeral segment events for benchmark harnesses.
type SegmentBuffer struct {
	Events []int
}

func (b *SegmentBuffer) Push(v int) {
	b.Events = append(b.Events, v)
}

func (b *SegmentBuffer) Snapshot() []int {
	out := append([]int(nil), b.Events...)
	sort.Ints(out)
	return out
}

func (b *SegmentBuffer) Reset() {
	b.Events = b.Events[:0]
}
