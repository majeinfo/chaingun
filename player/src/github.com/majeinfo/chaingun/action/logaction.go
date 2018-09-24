package action

import (
	log "github.com/sirupsen/logrus"	
   "github.com/majeinfo/chaingun/reporter"
   "github.com/majeinfo/chaingun/config"   
)

type LogAction struct {
	Message string	`yaml:"message"`
}

var (
	disable_log bool = false
)

func (s LogAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	if disable_log {
		return true
	}
	log.Infof("[LOG] %s", SubstParams(sessionMap, s.Message))
	return true
}

func NewLogAction(a map[interface{}]interface{}) (LogAction, bool) {
	return LogAction{a["message"].(string)}, true
}

func DisableAction(no_log bool) {
	disable_log = no_log
}