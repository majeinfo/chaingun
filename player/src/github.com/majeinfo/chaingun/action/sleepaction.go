package action

import (
	"time"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// SleepAction describes a sleep Action
type SleepAction struct {
	Duration int `yaml:"duration"` // in milli-seconds
}

// Execute a sleep Action
func (s SleepAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	time.Sleep(time.Duration(s.Duration) * time.Millisecond)
	return true
}

// NewSleepAction creates a new sleep Action
func NewSleepAction(a map[interface{}]interface{}) (SleepAction, bool) {
	valid := true
	if a["duration"] == nil {
		log.Error("sleep action needs 'duration' attribute")
		a["duration"] = 0
		valid = false
	}
	return SleepAction{a["duration"].(int)}, valid
}
