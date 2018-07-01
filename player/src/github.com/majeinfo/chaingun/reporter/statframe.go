package reporter

type StatFrame struct {
	Type    string `json:"type"`
	Time    int64  `json:"time"`
	Latency int    `json:"latency"`
	Reqs    int    `json:"reqs"`
}
