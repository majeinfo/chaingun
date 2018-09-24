package action

import (
   "github.com/majeinfo/chaingun/reporter"
   "github.com/majeinfo/chaingun/config"   
)

type TcpAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	Title string `yaml:"title"`
}

func (t TcpAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	DoTcpRequest(t, resultsChannel, sessionMap)
	return true
}

func NewTcpAction(a map[interface{}]interface{}) (TcpAction, bool) {

	// TODO validation
	return TcpAction{
		a["address"].(string),
		a["payload"].(string),
		a["title"].(string),
	}, true
}
