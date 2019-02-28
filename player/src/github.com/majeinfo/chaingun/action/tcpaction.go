package action

import (
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
)

// TCPAction describes a TCP Action
type TCPAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	Title   string `yaml:"title"`
}

// Execute a TCP Action
func (t TCPAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	DoTCPRequest(t, resultsChannel, sessionMap)
	return true
}

// NewTCPAction createsa new TCP Action
func NewTCPAction(a map[interface{}]interface{}) (TCPAction, bool) {
	// TODO validation
	return TCPAction{
		Address: a["address"].(string),
		Payload: a["payload"].(string),
		Title:   a["title"].(string),
	}, true
}
