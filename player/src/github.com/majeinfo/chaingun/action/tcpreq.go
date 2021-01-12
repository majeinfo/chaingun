package action

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

const (
	REPORTER_TCP string = "TCP"
)

// DoTCPRequest accepts a TcpAction and a one-way channel to write the results to.
func DoTCPRequest(tcpAction TCPAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry) {
	var payload string

	address := SubstParams(sessionMap, tcpAction.Address, vulog)
	if len(tcpAction.Payload_bytes) == 0 {
		payload = SubstParams(sessionMap, tcpAction.Payload, vulog)
	} else {
		payload = string(tcpAction.Payload_bytes)
	}

	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Errorf("TCP socket could not be created, error: %s", err)
		return
	}
	// conn.SetDeadline(time.Now().Add(100 * time.Millisecond))
	defer conn.Close()

	start := time.Now()

	if _, err := fmt.Fprintf(conn, payload); err != nil {
		log.Errorf("TCP request failed with error: %s", err)
	} else {
		// We do not read all the returned bytes
		buffer := make([]byte, 1000)
		reader := bufio.NewReader(conn)
		nbytes, _ := reader.Read(buffer)
		log.Debugf("%d bytes read !", nbytes)
	}

	elapsed := time.Since(start)
	resultsChannel <- buildTCPResult(sessionMap["UID"], 0, 200, elapsed.Nanoseconds(), tcpAction.Title)
}

func buildTCPResult(vid string, contentLength int, status int, elapsed int64, title string) reporter.SampleReqResult {
	sampleReqResult := reporter.SampleReqResult{
		Vid:     vid,
		Type:    REPORTER_TCP,
		Latency: elapsed,
		Size:    contentLength,
		Status:  status,
		Title:   title,
		When:    time.Since(reporter.SimulationStart).Nanoseconds(),
	}
	return sampleReqResult
}
