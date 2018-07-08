package action

import (
	"io/ioutil"
	"os"

	"github.com/majeinfo/chaingun/config"
	log "github.com/sirupsen/logrus"
)

func BuildActionList(playbook *config.TestDef) ([]Action, bool) {
	valid := true
	actions := make([]Action, len(playbook.Actions), len(playbook.Actions))
	for _, element := range playbook.Actions {
		log.Debugf("element=%v", element)
		for key, value := range element {
			var action Action
			actionMap := value.(map[interface{}]interface{})
			switch key {
			case "sleep":
				action = NewSleepAction(actionMap)
				break
			case "http":
				action = NewHttpAction(actionMap)
				break
			case "ws":
				action = NewWSAction(actionMap)
				break
			case "tcp":
				action = NewTcpAction(actionMap)
				break
			case "udp":
				action = NewUdpAction(actionMap)
				break
			default:
				valid = false
				log.Errorf("Unknown action type encountered: %s", key)
				break
			}
			if valid {
				actions = append(actions, action)
			}
		}
	}
	return actions, valid
}

func getBody(action map[interface{}]interface{}) string {
	//var body string = ""
	if action["body"] != nil {
		return action["body"].(string)
	} else {
		return ""
	}
}

func getTemplate(action map[interface{}]interface{}) string {
	if action["template"] != nil {
		var templateFile = action["template"].(string)
		dir, _ := os.Getwd()
		templateData, _ := ioutil.ReadFile(dir + "/templates/" + templateFile)
		return string(templateData)
	} else {
		return ""
	}
}
