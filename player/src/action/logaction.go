package action

import (
	"net/url"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// LogAction describes a log Action
type LogAction struct {
	Message string `yaml:"message"`
}

var (
	disableLog bool
)

// Execute a log Action
func (s LogAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, _ *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	if disableLog {
		return true
	}
	unesc, _ := url.QueryUnescape(SubstParams(sessionMap, s.Message, vulog))
	vulog.Infof("[LOG] %s", unesc)
	return true
}

// NewLogAction creates a new Log Action
func NewLogAction(a map[interface{}]interface{}) (LogAction, bool) {
	valid := true
	if a["message"] == nil {
		log.Error("log action needs 'message' attribute")
		a["message"] = ""
		valid = false
	}
	return LogAction{Message: a["message"].(string)}, valid
}

// DisableAction is called upon --no-log arg on command line
func DisableLogAction(noLog bool) {
	disableLog = noLog
}
