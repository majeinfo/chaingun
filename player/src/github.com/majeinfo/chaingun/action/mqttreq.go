package action

import (
	/*
		_ "bytes"
		_ "crypto/tls"
		_ "fmt"
	*/
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	/*
		_ "io/ioutil"
		_ "mime/multipart"
		_ "net/http"
		_ "path/filepath"
		_ "strconv"
		_ "strings"
		_ "time"
	*/
	//log "github.com/sirupsen/logrus"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// DoMQTTRequest accepts a MqttAction and a one-way channel to write the results to.
func DoMQTTRequest(mqttAction MQTTAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqttAction.URL)
	opts.SetUsername(mqttAction.Username)
	opts.SetPassword(mqttAction.Password)
	//opts.SetClientID(clientId)

	return true
}
