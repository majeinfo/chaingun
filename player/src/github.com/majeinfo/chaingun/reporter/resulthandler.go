package reporter

import (
	"encoding/json"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	stopNow   bool
	pvuCount  *int
	broadcast *chan []byte
)

/*
 * Starts the per second aggregator and then forwards any HttpRequestResult messages to it through the channel.
 */
func AcceptResults(resChannel chan HttpReqResult, vuCount *int, bcast *chan []byte) {
	log.Debug("AcceptResults")
	pvuCount = vuCount
	broadcast = bcast
	perSecondAggregatorChannel := make(chan *HttpReqResult, 500)
	stopNow = false
	go aggregatePerSecondHandler(perSecondAggregatorChannel)

	for {
		select {
		case msg := <-resChannel:
			perSecondAggregatorChannel <- &msg
			WriteResult(&msg) // sync write result to file for later processing.
			break
		case <-time.After(100 * time.Microsecond):
			break
			//		default:
			//			// This is troublesome. If too high, throughput is bad. Too low, CPU use goes up too much
			//			// Using a sync channel kills performance
			//			time.Sleep(100 * time.Microsecond)
		}
	}
}

// Stop the WS Server and the aggregator
func StopResults() {
	log.Debug("StopResults")
	stopNow = true
}

/**
 * Loops indefinitely. The inner loop runs for exactly one second before submitting its
 * results to the WebSocket handler, then the aggregates are reset and restarted.
 */
func aggregatePerSecondHandler(perSecondChannel chan *HttpReqResult) {

	for {
		var totalReq int
		var totalLatency int
		until := time.Now().UnixNano() + 1000000000
		for time.Now().UnixNano() < until {
			select {
			case msg := <-perSecondChannel:
				totalReq++
				totalLatency += int(msg.Latency / 1000) // measure in microseconds
			default:
				// Can be trouble. Uses too much CPU if low, limits throughput if too high
				time.Sleep(100 * time.Microsecond)
			}
		}
		// concurrently assemble the result and send it off to the websocket.
		go assembleAndSendResult(totalReq, totalLatency)

		if stopNow {
			break
		}
	}
}

// Total count of Requests cumulated by all VUs
var SuperTotalReq int

func assembleAndSendResult(totalReq int, totalLatency int) {
	avgLatency := 0
	if totalReq > 0 {
		SuperTotalReq += totalReq
		avgLatency = totalLatency / totalReq
	}
	statFrame := StatFrame{
		Type:    "rt",
		Time:    time.Since(SimulationStart).Nanoseconds() / 1000000000, // seconds
		Latency: avgLatency,                                             // microseconds
		Reqs:    totalReq,
	}

	log.Infof("Time: %d TotalReq: %d, VUCount: %d, Avg latency: %d μs (%d ms) req/s: %d", statFrame.Time, SuperTotalReq, *pvuCount, statFrame.Latency, statFrame.Latency/1000, statFrame.Reqs)

	serializedFrame, _ := json.Marshal(statFrame)
	*broadcast <- serializedFrame
}

// EOF
