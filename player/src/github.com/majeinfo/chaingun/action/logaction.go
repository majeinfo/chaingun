package action

import (
	log "github.com/sirupsen/logrus"	
   "github.com/majeinfo/chaingun/reporter"
   "github.com/majeinfo/chaingun/config"   
)

type LogAction struct {
	Message string	`yaml:"message"`
}

func (s LogAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	log.Infof("[LOG] %s", SubstParams(sessionMap, s.Message))
	return true
}

func NewLogAction(a map[interface{}]interface{}) LogAction {
	return LogAction{a["message"].(string)}
}
