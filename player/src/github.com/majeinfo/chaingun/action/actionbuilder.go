package action

import (
	"io/ioutil"
	"net/url"

	"github.com/majeinfo/chaingun/config"
	log "github.com/sirupsen/logrus"
)

var (
	gpScriptDir string
)

// BuildActionList builds all the Actions !
func BuildActionList(playbook *config.TestDef, scriptDir string) ([]Action, bool) {
	gpScriptDir = scriptDir
	valid := true

	actions := make([]Action, len(playbook.Actions), len(playbook.Actions))
	for _, element := range playbook.Actions {
		log.Debugf("element=%v", element)
		for key, value := range element {
			var action Action
			actionMap := value.(map[interface{}]interface{})
			switch key {
			case "sleep":
				action, valid = NewSleepAction(actionMap)
				break
			case "http":
				action, valid = NewHTTPAction(actionMap, playbook.DfltValues)
				break
			case "mqtt":
				action, valid = NewMQTTAction(actionMap, playbook.DfltValues)
				break
			case "ws":
				action, valid = NewWSAction(actionMap, playbook.DfltValues)
				break
			case "tcp":
				action, valid = NewTCPAction(actionMap)
				break
			case "udp":
				action, valid = NewUDPAction(actionMap)
				break
			case "log":
				action, valid = NewLogAction(actionMap)
				break
			case "setvar":
				action, valid = NewSetVarAction(actionMap)
			default:
				valid = false
				log.Errorf("Unknown action type encountered: %s", key)
				break
			}
			if valid {
				actions = append(actions, action)
			} else {
				return actions, valid
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

func getTemplate(action map[interface{}]interface{}) (string, bool) {
	if action["template"] != nil {
		var templateFile = action["template"].(string)
		log.Debugf("getTemplate: %s", templateFile)

		// TODO: should check if template file has been found - how to do in distributed mode ?
		if templateFile[0] != '/' {
			templateData, err := ioutil.ReadFile(gpScriptDir + "/" + templateFile)
			if err != nil {
				log.Errorf("Error while reading template file %s: %v", gpScriptDir+"/"+templateFile, err)
				return "", false
			}
			log.Debugf("templateData: %s", string(templateData))
			return string(templateData), true
		}

		templateData, err := ioutil.ReadFile(templateFile)
		if err != nil {
			log.Errorf("Error while reading template file %s: %v", templateFile, err)
			return "", false
		}
		log.Debugf("templateData: %s", string(templateData))
		return string(templateData), true

	}

	log.Debugf("no template data")
	return "", true
}

func getFileToUpload(filename string) ([]byte, bool) {
	log.Debugf("getFileToUpload: %s", filename)

	// TODO: should check if  file has been found - how to do in distributed mode ?
	if filename[0] != '/' {
		content, err := ioutil.ReadFile(gpScriptDir + "/" + filename)
		if err != nil {
			log.Errorf("Error while reading file %s: %v", gpScriptDir+"/"+filename, err)
			return nil, false
		}
		log.Debugf("content: %s", string(content))
		return content, true
	}

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf("Error while reading file %s: %v", filename, err)
		return nil, false
	}
	log.Debugf("content: %s", string(content))
	return content, true

}
