package action

import (
   "github.com/majeinfo/chaingun/reporter"
)

type UdpAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	Title string `yaml:"title"`
}

func (t UdpAction) Execute(resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string) {
	DoUdpRequest(t, resultsChannel, sessionMap)
}

func NewUdpAction(a map[interface{}]interface{}) UdpAction {

	// TODO validation
	return UdpAction{
		a["address"].(string),
		a["payload"].(string),
		a["title"].(string),
	}
}
