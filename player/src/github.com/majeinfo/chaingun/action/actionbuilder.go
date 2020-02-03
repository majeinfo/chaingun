package action

import (
	"io/ioutil"
	"net/url"

	"github.com/Knetic/govaluate"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

// BuildActionList builds all the Actions !
func BuildActionList(playbook *config.TestDef) ([]FullAction, []FullAction, bool) {
	var pre_actions []FullAction
	var actions []FullAction
	valid_pre_actions := true
	valid_actions := true

	pre_actions, valid_pre_actions = _buildActionList(playbook, playbook.PreActions)
	actions, valid_actions = _buildActionList(playbook, playbook.Actions)

	return pre_actions, actions, valid_pre_actions && valid_actions
}

func _buildActionList(playbook *config.TestDef, playbook_actions []map[string]interface{}) ([]FullAction, bool) {
	valid := true
	actions := make([]FullAction, len(playbook_actions), len(playbook_actions))

	for _, element := range playbook_actions {
		log.Debugf("element=%v", element)
		var action Action
		var fullAction FullAction

		for key, value := range element {
			log.Debugf("key=%s, value=%v", key, value)
			if key == "when" {
				var whenErr error
				fullAction.When = value.(string)
				fullAction.CompiledWhen, whenErr = govaluate.NewEvaluableExpressionWithFunctions(value.(string), getExpressionFunctions())
				if whenErr != nil {
					log.Errorf("When Expression '%s' cannot be compiled (%s)", fullAction.When, whenErr)
					valid = false
				}
			} else {
				var actionMap map[interface{}]interface{}
				var ok bool
				if actionMap, ok = value.(map[interface{}]interface{}); !ok {
					log.Errorf("Either %s is not allowed here, either its value is not a subdocument", key)
					valid = false
					continue
				}
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
				case "assert":
					action, valid = NewAssertAction(actionMap)
				default:
					valid = false
					log.Errorf("Unknown action type encountered: %s", key)
					break
				}
			}
		}
		if valid {
			fullAction.Action = action
			actions = append(actions, fullAction)
		} else {
			return actions, valid
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

func getBody(action map[interface{}]interface{}) (string, bool) {
	//var body string = ""
	if action["body"] != nil {
		// Fix bug: if the user wants to send JSON, he may forget to enclose the body between quotes
		// 				body: {k:v}
		// and the confuses the YAML analyzer...
		// So we check the type of the "body" value
		switch v := action["body"].(type) {
		case string:
		default:
			log.Errorf("body value is not of type string: %v", v)
			return "", false
		}
		return action["body"].(string), true
	}

	return "", true
}

func getTemplate(action map[interface{}]interface{}) (string, bool) {
	if action["template"] != nil {
		var templateFile = action["template"].(string)
		log.Debugf("getTemplate: %s", templateFile)

		templateFile = utils.ComputeFilename(templateFile, gpScriptDir)

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
	filename = utils.ComputeFilename(filename, gpScriptDir)

	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Errorf("Error while reading file %s: %v", filename, err)
		return nil, false
	}
	log.Debugf("content: %s", string(content))
	return content, true

}
