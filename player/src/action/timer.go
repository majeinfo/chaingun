package action

import (
	"strconv"
	"time"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

const (
	REPORTER_TIMER string = "TIMER"
	timer_prefix = "__timer__"
)

type StartTimerAction struct {
	Name string `yaml:"name"`
}

type EndTimerAction struct {
	Name string `yaml:"name"`
}

// Execute a Start Timer Action
func (s StartTimerAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, _ *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	// Check timer not already started
	if _, ok := sessionMap[s.Name]; ok {
		vulog.Errorf("Time %s already started and not stopped", s.Name)
		return false
	}

	//sessionMap[s.Name] = strconv.ParseInt(time.Now().UnixNano(), 10, 64)
	sessionMap[s.Name] = strconv.FormatInt(int64(time.Now().UnixNano()), 10)
	vulog.Debugf("Timer %s started at %s", s.Name, sessionMap[s.Name])
	return true
}

// Execute a End Timer Action
func (s EndTimerAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, _ *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	// Check the Timer has been started
	if startTime, ok := sessionMap[s.Name]; ok {
		start, _ := strconv.ParseInt(startTime, 10, 64)
		delta := time.Now().UnixNano() - start
		vulog.Debugf("Timer %s measures %v", s.Name, delta)

		sampleReqResult := reporter.SampleReqResult{
			Vid:     sessionMap["UID"],
			Type:    REPORTER_TIMER,
			Latency: delta,
			Size:    0,
			Status:  200,
			Title:   s.Name,
			When:    time.Since(reporter.SimulationStart).Nanoseconds(),
		}
		resultsChannel <- sampleReqResult

		return true
	}

	// Error...
	vulog.Errorf("Missing start_timer action for timer %s", s.Name)
	return false
}

// NewStartTimerAction creates a new Start Timer Action
func NewStartTimerAction(a map[interface{}]interface{}) (StartTimerAction, bool) {
	valid := true
	if a["name"] == nil {
		log.Error("start timer action needs 'name' attribute")
		a["name"] = ""
		valid = false
	}
	return StartTimerAction{Name: timer_prefix + a["name"].(string)}, valid
}

// NewStartTimerAction creates a new Start Timer Action
func NewEndTimerAction(a map[interface{}]interface{}) (EndTimerAction, bool) {
	valid := true
	if a["name"] == nil {
		log.Error("end timer action needs 'name' attribute")
		a["name"] = ""
		valid = false
	}
	return EndTimerAction{Name: timer_prefix + a["name"].(string)}, valid
}


