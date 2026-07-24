package r2p

import "fmt"

// StagingUtil formats staging paths for support tooling.
type StagingUtil struct {
	Root string
}

func (s *StagingUtil) Path(unit string, slot int) string {
	return fmt.Sprintf("%s/%s_%d.stage", s.Root, unit, slot)
}

func (s *StagingUtil) List(units []string) []string {
	out := make([]string, 0, len(units))
	for i, u := range units {
		out = append(out, s.Path(u, i))
	}
	return out
}
