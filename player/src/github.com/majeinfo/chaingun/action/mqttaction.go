package action

import (
	/*
		_ "errors"
		_ "strings"
	*/
	//mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// MQTTAction describes a MQTT Action
type MQTTAction struct {
	URL      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Execute a MQTT Action
func (h MQTTAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	return DoMQTTRequest(h, resultsChannel, sessionMap, playbook)
}

// NewMQTTAction creates a new MQTT Action
func NewMQTTAction(a map[interface{}]interface{}, dflt config.Default) (MQTTAction, bool) {
	log.Debugf("NewMqttAction=%v", a)
	valid := true

	if a["url"] == "" || a["url"] == nil {
		log.Error("MqttAction must define a URL.")
		valid = false
	} else {
		valid = setDefaultURL(a, dflt)
	}

	if a["username"] == nil {
		a["username"] = ""
	}
	if a["password"] == nil {
		a["password"] = ""
	}

	mqttAction := MQTTAction{
		URL:      a["url"].(string),
		Username: a["username"].(string),
		Password: a["password"].(string),
	}

	log.Debugf("MQTTAction: %v", mqttAction)

	return mqttAction, valid
}
