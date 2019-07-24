package action

import (
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// TCPAction describes a TCP Action
type TCPAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	Title   string `yaml:"title"`
}

// Execute a TCP Action
func (t TCPAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	DoTCPRequest(t, resultsChannel, sessionMap, vulog)
	return true
}

// NewTCPAction createsa new TCP Action
func NewTCPAction(a map[interface{}]interface{}) (TCPAction, bool) {
	valid := true

	if a["title"] == "" || a["title"] == nil {
		log.Error("TCPAction must define a title.")
		a["title"] = ""
		valid = false
	}
	if a["address"] == "" || a["address"] == nil {
		log.Error("TCPAction must define a target address.")
		a["address"] = ""
		valid = false
	}
	if a["payload"] == nil {
		a["payload"] = ""
	}

	tcpAction := TCPAction{
		Address: a["address"].(string),
		Payload: a["payload"].(string),
		Title:   a["title"].(string),
	}

	log.Debugf("TCPAction: %v", tcpAction)

	return tcpAction, valid
}
