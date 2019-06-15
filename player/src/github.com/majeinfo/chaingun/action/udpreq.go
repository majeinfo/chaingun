package action

import (
    "fmt"
    "github.com/majeinfo/chaingun/reporter"
    log "github.com/sirupsen/logrus"
    "net"
    "time"
)

var udpconn *net.UDPConn

// DoUDPRequest accepts a UdpAction and a one-way channel to write the results to.
func DoUDPRequest(udpAction UDPAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry) {

    address := SubstParams(sessionMap, udpAction.Address, vulog)
    payload := SubstParams(sessionMap, udpAction.Payload, vulog)

    if udpconn == nil {
        ServerAddr, err := net.ResolveUDPAddr("udp", address) //"127.0.0.1:10001")
        if err != nil {
            log.Errorf("Error ResolveUDPAddr remote: %s", err.Error())
        }

        LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
        if err != nil {
            log.Errorf("Error ResolveUDPAddr local: %s", err.Error())
        }

        udpconn, err = net.DialUDP("udp", LocalAddr, ServerAddr)
        if err != nil {
            log.Errorf("Error Dial: %s", err.Error())
        }
    }
    //defer Conn.Close()
    start := time.Now()
    if udpconn != nil {
        _, err := fmt.Fprintf(udpconn, payload+"\r\n")
        if err != nil {
            log.Errorf("UDP request failed with error: %s", err)
            udpconn = nil
        }
    }

    elapsed := time.Since(start)
    resultsChannel <- buildUDPResult(sessionMap["UID"], 0, 200, elapsed.Nanoseconds(), udpAction.Title)

}

func buildUDPResult(vid string, contentLength int, status int, elapsed int64, title string) reporter.SampleReqResult {
    sampleReqResult := reporter.SampleReqResult{
        Vid:     vid,
        Type:    "UDP",
        Latency: elapsed,
        Size:    contentLength,
        Status:  status,
        Title:   title,
        When:    time.Since(reporter.SimulationStart).Nanoseconds(),
    }
    return sampleReqResult
}
