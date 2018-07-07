package action

import (
   "time"
   "github.com/majeinfo/chaingun/reporter"
   "github.com/majeinfo/chaingun/config"   
)

type SleepAction struct {
	Duration int `yaml:"duration"`
}

func (s SleepAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	time.Sleep(time.Duration(s.Duration) * time.Second)
	return true
}

func NewSleepAction(a map[interface{}]interface{}) SleepAction {
	return SleepAction{a["duration"].(int)}
}
