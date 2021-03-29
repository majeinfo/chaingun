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
	REPORTER_UDP string = "UDP"
)

// DoUDPRequest accepts a UdpAction and a one-way channel to write the results to.
func DoUDPRequest(udpAction UDPAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, _ *config.VUContext, vulog *log.Entry) {
	var payload string

	address := SubstParams(sessionMap, udpAction.Address, vulog)
	if len(udpAction.Payload_bytes) == 0 {
		payload = SubstParams(sessionMap, udpAction.Payload, vulog)
	} else {
		payload = string(udpAction.Payload_bytes)
	}

	conn, err := net.Dial("udp", address)
	if err != nil {
		log.Errorf("UDP socket could not be created, error: %s", err.Error())
		return
	}
	// conn.SetDeadline(time.Now().Add(100 * time.Millisecond))

	defer conn.Close()
	start := time.Now()

	if _, err := fmt.Fprintf(conn, payload); err != nil {
		log.Errorf("UDP request failed with error: %s", err)
	} else {
		// We do not read all the returned bytes
		buffer := make([]byte, 1000)
		reader := bufio.NewReader(conn)
		reader.Read(buffer)
	}

	elapsed := time.Since(start)
	resultsChannel <- buildUDPResult(sessionMap["UID"], 0, 200, elapsed.Nanoseconds(), udpAction.Title)
}

func buildUDPResult(vid string, contentLength int, status int, elapsed int64, title string) reporter.SampleReqResult {
	sampleReqResult := reporter.SampleReqResult{
		Vid:     vid,
		Type:    REPORTER_UDP,
		Latency: elapsed,
		Size:    contentLength,
		Status:  status,
		Title:   title,
		When:    time.Since(reporter.SimulationStart).Nanoseconds(),
	}
	return sampleReqResult
}
