package reporter

const (
	NETWORK_ERROR = -1
)

// SampleReqResult describes the structure of sampling result
// if Type value is 'GLOBAL', the structure is not a request sample
// but a global result which indicates in Size field how many VU are handled.
// In that case, all the other fields are empty or equal 0.
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
