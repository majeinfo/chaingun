package action

import (
   "github.com/majeinfo/chaingun/reporter"
)

type TcpAction struct {
	Address string `yaml:"address"`
	Payload string `yaml:"payload"`
	Title string `yaml:"title"`
}

func (t TcpAction) Execute(resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string) {
	DoTcpRequest(t, resultsChannel, sessionMap)
}

func NewTcpAction(a map[interface{}]interface{}) TcpAction {

	// TODO validation
	return TcpAction{
		a["address"].(string),
		a["payload"].(string),
		a["title"].(string),
	}
}
