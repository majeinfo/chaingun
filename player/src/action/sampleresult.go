package action

import (
	"time"

	"github.com/majeinfo/chaingun/reporter"
)

func buildSampleResult(actionType string, vid string, contentLength int, status int, elapsed int64, title string, fullreq string) reporter.SampleReqResult {
	sampleReqResult := reporter.SampleReqResult{
		Vid:         vid,
		Type:        actionType,
		Latency:     elapsed,
		Size:        contentLength,
		Status:      status,
		Title:       title,
		When:        time.Since(reporter.SimulationStart).Nanoseconds(),
		FullRequest: fullreq,
	}
	return sampleReqResult
}

func updateWhenTime(reqResult *reporter.SampleReqResult) {
	reqResult.When = time.Since(reporter.SimulationStart).Nanoseconds()
}