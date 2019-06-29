package reporter

import (
	"encoding/json"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	stopNow        bool
	lock_stopNow   sync.Mutex
	pvuCount       *int
	plock_vu_count *sync.Mutex
	broadcast      *chan []byte
)

/*
 * Starts the per second aggregator and then forwards any HttpRequestResult messages to it through the channel.
 */
func AcceptResults(resChannel chan SampleReqResult, vuCount *int, lock_vu_count *sync.Mutex, bcast *chan []byte, must_bcast bool) {
	log.Debug("AcceptResults called")
	pvuCount = vuCount
	plock_vu_count = lock_vu_count
	if must_bcast {
		broadcast = bcast
	}
	perSecondAggregatorChannel := make(chan *SampleReqResult, 500)
	lock_stopNow.Lock()
	stopNow = false
	lock_stopNow.Unlock()
	go aggregatePerSecondHandler(perSecondAggregatorChannel)

	for {
		// +build !trace
		lock_stopNow.Lock()
		if stopNow {
			lock_stopNow.Unlock()
			break
		}
		lock_stopNow.Unlock()
		// +build trace

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
	log.Debug("exit AcceptResults")
}

// Stop the WS Server and the aggregator
func StopResults() {
	log.Debug("StopResults")
	time.Sleep(2 * time.Second) // Give a chance to write down the last results before leaving
	// +build !trace
	lock_stopNow.Lock()
	stopNow = true
	lock_stopNow.Unlock()
	// +build trace
}

/**
 * Loops indefinitely. The inner loop runs for exactly one second before submitting its
 * results to the WebSocket handler, then the aggregates are reset and restarted.
 */
func aggregatePerSecondHandler(perSecondChannel chan *SampleReqResult) {
	log.Debug("aggregatePerSecondHandler called")
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

		lock_stopNow.Lock()
		if stopNow {
			lock_stopNow.Unlock()
			break
		}
		lock_stopNow.Unlock()
	}
	log.Debug("exit aggregatePerSecondHandler")
}

// Total count of Requests cumulated by all VUs
var SuperTotalReq int
var lock2 sync.Mutex

func assembleAndSendResult(totalReq int, totalLatency int) {
	log.Debug("assembleAndSendResult called")
	avgLatency := 0
	if totalReq > 0 {
		lock2.Lock() // Added to avoid race condition on SuperTotalReq
		SuperTotalReq += totalReq
		lock2.Unlock()
		avgLatency = totalLatency / totalReq
	}
	statFrame := StatFrame{
		Type:    "rt",
		Time:    time.Since(SimulationStart).Nanoseconds() / 1000000000, // seconds
		Latency: avgLatency,                                             // microseconds
		Reqs:    totalReq,
	}

	// +build !trace
	plock_vu_count.Lock()
	lock2.Lock()
	log.Infof("Time: %d TotalReq: %d, VUCount: %d, Avg latency: %d μs (%d ms) req/s: %d", statFrame.Time, SuperTotalReq, *pvuCount, statFrame.Latency, statFrame.Latency/1000, statFrame.Reqs)
	lock2.Unlock()
	plock_vu_count.Unlock()
	// +build trace

	if broadcast != nil {
		serializedFrame, _ := json.Marshal(statFrame)
		*broadcast <- serializedFrame
	}
	log.Debug("exit assembleAndSendResult")
}

// EOF
