package action

import (
   "github.com/majeinfo/chaingun/reporter"
   "github.com/majeinfo/chaingun/config"   
)

type UdpAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	Title string `yaml:"title"`
}

func (t UdpAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	DoUdpRequest(t, resultsChannel, sessionMap)
	return true
}

func NewUdpAction(a map[interface{}]interface{}) UdpAction {

	// TODO validation
	return UdpAction{
		a["address"].(string),
		a["payload"].(string),
		a["title"].(string),
	}
}
