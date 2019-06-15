package action

import (
	/*
		_ "errors"
		_ "strings"
	*/
	"crypto/tls"

	"time"
	//mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// MQTTAction describes a MQTT Action
type MQTTAction struct {
	URL                  string        `yaml:"url"`
	Title                string        `yaml:"title"`
	Username             string        `yaml:"username"`
	Password             string        `yaml:"password"`
	ClientID             string        `yaml:"clientid"`
	MaxReconnectInterval time.Duration `yaml:"maxreconnectinterval"`
	KeepAlive            time.Duration `yaml:"keppalive"`
	Payload              string        `yaml:"payload"`
	Qos                  byte          `yaml:"qos"`
	Topic                string        `yaml:"topic"`
	CertificatePath      string        `yaml:"certificatepath"`
	PrivateKeyPath       string        `yaml:"privatekeypath"`
	TLSConfig            *tls.Config
}

// Execute a MQTT Action
func (h MQTTAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	return DoMQTTRequest(h, resultsChannel, sessionMap, vulog, playbook)
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

	if a["clientid"] == "" || a["clientid"] == nil {
		a["clientid"] = "chaingun-by-JD"
	}

	if a["title"] == nil || a["title"] == "" {
		log.Error("MqttAction must define a title.")
		a["title"] = ""
		valid = false
	}

	if a["topic"] == "" || a["topic"] == nil {
		log.Error("MqttAction must define a Topic.")
		a["topic"] = ""
		valid = false
	}

	if a["payload"] == "" || a["payload"] == nil {
		log.Error("MqttAction must define a Payload.")
		a["payload"] = ""
		valid = false
	}

	if a["qos"] == "" || a["qos"] == nil {
		a["qos"] = 0
	}

	if a["username"] == nil {
		a["username"] = ""
	}
	if a["password"] == nil {
		a["password"] = ""
	}

	tlsconfig := tls.Config{}
	if a["certificatepath"] != nil && a["privatekeypath"] != nil {
		cer, err := tls.LoadX509KeyPair(a["certificatepath"].(string), a["privatekeypath"].(string))
		if err != nil {
			log.Errorf("Could not load certificate or private key: %v", err)
			valid = false
		} else {
			tlsconfig = tls.Config{Certificates: []tls.Certificate{cer}}
			tlsconfig.InsecureSkipVerify = true
		}
	}

	if !valid {
		log.Errorf("Your YAML Playbook contains an invalid MQTTAction, see errors listed above.")
		valid = false
	}

	mqttAction := MQTTAction{
		Title:                a["title"].(string),
		URL:                  a["url"].(string),
		ClientID:             a["clientid"].(string),
		Username:             a["username"].(string),
		Password:             a["password"].(string),
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            30 * time.Second,
		Payload:              a["payload"].(string),
		Qos:                  byte(a["qos"].(int)),
		Topic:                a["topic"].(string),
		TLSConfig:            &tlsconfig,
	}

	log.Debugf("MQTTAction: %v", mqttAction)

	return mqttAction, valid
}
