package action

import (
    "net"
    "fmt"
    "time"
    log "github.com/sirupsen/logrus"
    "github.com/majeinfo/chaingun/reporter"
)

var conn net.Conn
// Accepts a TcpAction and a one-way channel to write the results to.
func DoTcpRequest(tcpAction TcpAction, resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string) {

    address := SubstParams(sessionMap, tcpAction.Address)
    payload := SubstParams(sessionMap, tcpAction.Payload)

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

    _, err := fmt.Fprintf(conn, payload + "\r\n")
    if err != nil {
        log.Errorf("TCP request failed with error: %s", err)
        conn = nil
    }

    elapsed := time.Since(start)
    resultsChannel <- buildTcpResult(sessionMap["UID"], 0, 200, elapsed.Nanoseconds(), tcpAction.Title)
}

func buildTcpResult(vid string, contentLength int, status int, elapsed int64, title string) (reporter.HttpReqResult){
    httpReqResult := reporter.HttpReqResult {
		vid,
        "TCP",
        elapsed,
        contentLength,
        status,
        title,
        time.Since(reporter.SimulationStart).Nanoseconds(),
    }
    return httpReqResult
}
