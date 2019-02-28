package action

import (
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
)

// UDPAction describes a UDP Action
type UDPAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	Title   string `yaml:"title"`
}

// Execute an UDP Request
func (t UDPAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	DoUDPRequest(t, resultsChannel, sessionMap)
	return true
}

// NewUDPAction creates a new UDP Action
func NewUDPAction(a map[interface{}]interface{}) (UDPAction, bool) {
	// TODO validation
	return UDPAction{
		Address: a["address"].(string),
		Payload: a["payload"].(string),
		Title:   a["title"].(string),
	}, true
}
