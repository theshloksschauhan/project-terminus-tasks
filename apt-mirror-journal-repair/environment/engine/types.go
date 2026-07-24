package engine

type RowVec struct {
	UnitID    string `json:"unit_id"`
	Lag       int    `json:"lag"`
	SchedRC   bool   `json:"sched_rc"`
	LockRC    bool   `json:"lock_rc"`
	UnitRC    bool   `json:"unit_rc"`
	FacetHex  string `json:"facet_hex"`
	DriftCode int    `json:"drift_code"`
}

type Summary struct {
	SyncLabel string `json:"sync_label"`
	RunDigest string `json:"run_digest"`
	Width     int    `json:"width"`
}

type RunReport struct {
	Rows    []RowVec `json:"rows"`
	Summary Summary  `json:"summary"`
}

type LaneCtx struct {
	StampA   uint32
	StampB   uint32
	Bitmap   uint32
	Seq      int
	Validated bool
}

type Witness struct {
	Bytes string
	Seq   int
}
