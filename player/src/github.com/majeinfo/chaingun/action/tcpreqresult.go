package action

// TCPReqResult describes a TCP or UDP Result
type TCPReqResult struct {
    Type    string
    Latency int64
    Size    int
    Status  int
    Title   string
    When    int64
}
