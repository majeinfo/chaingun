package action

import (
	"encoding/base64"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// UDPAction describes a UDP Action
type UDPAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	//Payload64 string `yaml:"payload64"`
	Payload_bytes []byte
	Title   string `yaml:"title"`
}

// Execute an UDP Request
func (t UDPAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	DoUDPRequest(t, resultsChannel, sessionMap, vucontext, vulog)
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
	if a["payload"] != nil && a["payload64"] != nil {
		log.Error("Either payload or payload64 can be defined in a UDPAction.")
		a["payload"] = ""
		a["payload_bytes"] = []byte{}
		valid = false
	}
	if a["payload"] == nil {
		a["payload"] = ""
	}
	if a["payload64"] == nil {
		a["payload_bytes"] = []byte{}
	} else {
		data, err := base64.StdEncoding.DecodeString(a["payload64"].(string))
		if err != nil {
			log.Errorf("Error while decoding payload64 value: %s (%s)", a["payload64"], err.Error())
			a["payload_bytes"] = []byte{}
			valid = false
		} else {
			a["payload_bytes"] = data
		}
	}

	udpAction := UDPAction{
		Address: a["address"].(string),
		Payload: a["payload"].(string),
		Payload_bytes: a["payload_bytes"].([]byte),
		Title:   a["title"].(string),
	}

	log.Debugf("UDPAction: %v", udpAction)

	return udpAction, valid
}
