package action

import (
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

type LogAction struct {
	Message string `yaml:"message"`
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
	valid := true
	if a["message"] == nil {
		log.Error("log action needs 'message' attribute")
		a["message"] = ""
		valid = false
	}
	return LogAction{a["message"].(string)}, valid
}

// Called upon --no-log arg on command line
func DisableAction(no_log bool) {
	disable_log = no_log
}
