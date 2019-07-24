package action

import (
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// UDPAction describes a UDP Action
type UDPAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	Title   string `yaml:"title"`
}

// Execute an UDP Request
func (t UDPAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	DoUDPRequest(t, resultsChannel, sessionMap, vulog)
	return true
}

// NewUDPAction creates a new UDP Action
func NewUDPAction(a map[interface{}]interface{}) (UDPAction, bool) {
	valid := true

	if a["title"] == "" || a["title"] == nil {
		log.Error("UDPAction must define a title.")
		a["title"] = ""
		valid = false
	}
	if a["address"] == "" || a["address"] == nil {
		log.Error("UDPAction must define a target address.")
		a["address"] = ""
		valid = false
	}
	if a["payload"] == nil {
		a["payload"] = ""
	}

	udpAction := UDPAction{
		Address: a["address"].(string),
		Payload: a["payload"].(string),
		Title:   a["title"].(string),
	}

	log.Debugf("UDPAction: %v", udpAction)

	return udpAction, valid
}
