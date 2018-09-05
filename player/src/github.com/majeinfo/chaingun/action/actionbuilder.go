package action

import (
	"io/ioutil"
	"net/url"

	"github.com/majeinfo/chaingun/config"
	log "github.com/sirupsen/logrus"
)

var (
	gp_script_dir string
)

func BuildActionList(playbook *config.TestDef, script_dir string) ([]Action, bool) {
	gp_script_dir = script_dir
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
				action = NewHttpAction(actionMap, playbook.DfltValues)
				break
			case "ws":
				action = NewWSAction(actionMap, playbook.DfltValues)
				break
			case "tcp":
				action = NewTcpAction(actionMap)
				break
			case "udp":
				action = NewUdpAction(actionMap)
				break
			case "log":
				action = NewLogAction(actionMap)
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

// a["url"] must exists and not be nil or empty
func setDefaultURL(a map[interface{}]interface{}, dflt config.Default) bool {
	valid := true

	u, err := url.Parse(a["url"].(string))
	if err != nil {
		log.Errorf("Wrong URL: %s", err)
		valid = false
	} 
	if u.Scheme == "" {
		if dflt.Protocol == "" {
			log.Errorf("Scheme (protocol) missing for URL: %s", a["url"])
			valid = false
		} else {
			u.Scheme = dflt.Protocol
			log.Debugf("Use default protocol: %s", u.Scheme)
		}
	}
	if u.Host == "" {
		if dflt.Server == "" {
			log.Errorf("Host missing for URL: %s", a["url"])
			valid = false			
		} else {
			u.Host = dflt.Server
			log.Debugf("Use default server: %s", u.Host)
		}
	}

	// The "Path" value must be added unescaped because it can contains variables (like ${...}) 
	a["url"] = u.String()

	return valid
}

func getBody(action map[interface{}]interface{}) string {
	//var body string = ""
	if action["body"] != nil {
		return action["body"].(string)
	}

	return ""
}

func getTemplate(action map[interface{}]interface{}) string {
	if action["template"] != nil {
		var templateFile = action["template"].(string)
		log.Debugf("getTemplate: %s", templateFile)

		if templateFile[0] != '/' {
			templateData, _ := ioutil.ReadFile(gp_script_dir + "/" + templateFile)
			log.Debugf("templateData: %s", string(templateData))
			return string(templateData)
		} else {
			templateData, _ := ioutil.ReadFile(templateFile)
			log.Debugf("templateData: %s", string(templateData))
			return string(templateData)
		}
	}

	log.Debugf("no template data")
	return ""
}
