package action

import (
    "net"
    "fmt"
    "time"
    log "github.com/sirupsen/logrus"    
    "github.com/majeinfo/chaingun/reporter"
)

var udpconn *net.UDPConn

// Accepts a UdpAction and a one-way channel to write the results to.
func DoUdpRequest(udpAction UdpAction, resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string) {

    address := SubstParams(sessionMap, udpAction.Address)
    payload := SubstParams(sessionMap, udpAction.Payload)

    if udpconn == nil {
        ServerAddr,err := net.ResolveUDPAddr("udp", address) //"127.0.0.1:10001")
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
        _, err := fmt.Fprintf(udpconn, payload + "\r\n")
        if err != nil {
            log.Errorf("UDP request failed with error: %s", err)
            udpconn = nil
        }
    }

    elapsed := time.Since(start)
    resultsChannel <- buildUdpResult(sessionMap["UID"], 0, 200, elapsed.Nanoseconds(), udpAction.Title)

}

func buildUdpResult(vid string, contentLength int, status int, elapsed int64, title string) (reporter.HttpReqResult){
    httpReqResult := reporter.HttpReqResult {
		vid,
        "UDP",
        elapsed,
        contentLength,
        status,
        title,
        time.Since(reporter.SimulationStart).Nanoseconds(),
    }
    return httpReqResult
}
