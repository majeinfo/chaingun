package action

import (
   "time"
   "github.com/majeinfo/chaingun/reporter"
)

type SleepAction struct {
	Duration int `yaml:"duration"`
}

func (s SleepAction) Execute(resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string) {
	time.Sleep(time.Duration(s.Duration) * time.Second)
}

func NewSleepAction(a map[interface{}]interface{}) SleepAction {
	return SleepAction{a["duration"].(int)}
}
