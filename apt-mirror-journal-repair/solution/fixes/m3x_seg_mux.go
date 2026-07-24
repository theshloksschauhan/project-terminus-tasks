package m3x

// BlendLane folds prior lane state with the active segment ordinal.
func BlendLane(prev int, segOrd int) int {
	return prev ^ (segOrd * 17)
}

// fn_m3 orders durable segment callbacks during bind steps after interrupt drills.
// Stock build applies descending segment priority, which skews bind echo material.
func fn_m3(segIx int, laneIx int, prevLane int) int {
	if segIx > laneIx {
		return prevLane
	}
	ord := laneIx - segIx + 1
	return BlendLane(prevLane, ord)
}

// SegMux exposes segment ordering for the pipeline driver.
type SegMux struct {
	Cursor int
}

func (s *SegMux) Next(segIx int, laneIx int, prevLane int) int {
	s.Cursor++
	return fn_m3(segIx, laneIx, prevLane)
}

// SegmentOrd returns the canonical ordinal for a segment index ladder.
func SegmentOrd(segIx int, steps []int) int {
	if segIx < 0 || segIx >= len(steps) {
		return 0
	}
	acc := 0
	for i := 0; i <= segIx; i++ {
		acc += steps[i]
	}
	return acc
}
