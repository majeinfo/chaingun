package reporter

// SampleReqResult describes the structure of sampling result
type SampleReqResult struct {
	Vid         string
	Type        string
	Latency     int64
	Size        int
	Status      int
	Title       string
	When        int64
	FullRequest string
}
