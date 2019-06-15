package action

import (
    "fmt"
    "github.com/majeinfo/chaingun/reporter"
    log "github.com/sirupsen/logrus"
    "net"
    "time"
)

var conn net.Conn

// DoTCPRequest accepts a TcpAction and a one-way channel to write the results to.
func DoTCPRequest(tcpAction TCPAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry) {

    address := SubstParams(sessionMap, tcpAction.Address, vulog)
    payload := SubstParams(sessionMap, tcpAction.Payload, vulog)

    if conn == nil {

        _, err := net.Dial("tcp", address)
        if err != nil {
            log.Errorf("TCP socket closed, error: %s", err)
            conn = nil
            return
        }
        // conn.SetDeadline(time.Now().Add(100 * time.Millisecond))
    }

    start := time.Now()

    _, err := fmt.Fprintf(conn, payload+"\r\n")
    if err != nil {
        log.Errorf("TCP request failed with error: %s", err)
        conn = nil
    }

    elapsed := time.Since(start)
    resultsChannel <- buildTCPResult(sessionMap["UID"], 0, 200, elapsed.Nanoseconds(), tcpAction.Title)
}

func buildTCPResult(vid string, contentLength int, status int, elapsed int64, title string) reporter.SampleReqResult {
    sampleReqResult := reporter.SampleReqResult{
        Vid:     vid,
        Type:    "TCP",
        Latency: elapsed,
        Size:    contentLength,
        Status:  status,
        Title:   title,
        When:    time.Since(reporter.SimulationStart).Nanoseconds(),
    }
    return sampleReqResult
}
