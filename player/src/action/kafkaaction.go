package action

import (
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// KafkaAction describes a MongoDB Action
type KafkaAction struct {
	Brokers          []string          `yaml:"brokers"`
	Title            string            `yaml:"title"`
	Topic            string            `yaml:"topic"`
	TLSEnabled       bool              `yaml:"tlsEnabled"`
	Command          string            `yaml:"command"`
	Key              string            `yaml:"key"`
	Value            string            `yaml:"value"`
	ResponseHandlers []ResponseHandler `yaml:"responses"`
}

// Execute a Kafka Action
func (h KafkaAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	vulog.Data["action"] = h.Title
	return DoKafkaRequest(h, resultsChannel, sessionMap, vucontext, vulog, playbook)
}

// NewKafkaAction creates a new Kafka Action
func NewKafkaAction(a map[interface{}]interface{}, dflt config.Default, playbook *config.TestDef) (KafkaAction, bool) {
	log.Debugf("NewKafkaAction=%v", a)
	valid := true

	if a["brokers"] == nil || a["brokers"] == "" {
		log.Error("KafkaAction must define brokers (at leat one !)")
		a["brokers"] = []string{""}
		valid = false
	}

	if a["title"] == nil || a["title"] == "" {
		log.Error("KafkaAction must define a title")
		a["title"] = ""
		valid = false
	}

	if a["topic"] == nil || a["topic"] == "" {
		log.Error("KafkaAction must define a topic")
		a["topic"] = ""
		valid = false
	}

	if a["command"] == nil || a["command"] == "" {
		log.Error("KafkaAction must have a command")
		a["command"] = ""
		valid = false
	} else if _, err := config.IsValidKafkaCommand(a["command"].(string)); err != nil {
		log.Error("%s", err)
		valid = false
	}

	if a["key"] == nil {
		a["key"] = ""
	}

	if a["value"] == nil {
		a["value"] = ""
	}

	if a["command"] == "write" {
		if a["value"] == "" {
			log.Error("KafkaAction must define a value for write action")
			valid = false
		}
	}

	responseHandlers, validResp := NewResponseHandlers(a)

	if !valid || !validResp {
		log.Errorf("Your YAML Playbook contains an invalid KafkaAction, see errors listed above")
		valid = false
	}

	kafkaAction := KafkaAction{
		Brokers:          []string{"localhost:29092", "localhost:29092"},
		Topic:            a["topic"].(string),
		Title:            a["title"].(string),
		TLSEnabled:       false,
		Command:          a["command"].(string),
		Key:              a["key"].(string),
		Value:            a["value"].(string),
		ResponseHandlers: responseHandlers,
	}

	log.Debugf("KafkaAction: %v", kafkaAction)

	return kafkaAction, valid
}