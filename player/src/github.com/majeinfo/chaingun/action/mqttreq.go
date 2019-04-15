package action

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
	"time"
)

/*
type MQTTLogger struct{}

func (MQTTLogger) Println(v ...interface{})               { fmt.Println(v) }
func (MQTTLogger) Printf(format string, v ...interface{}) { fmt.Print(format, v) }
*/

// DoMQTTRequest accepts a MqttAction and a one-way channel to write the results to.
func DoMQTTRequest(mqttAction MQTTAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	//MQTT.DEBUG = MQTTLogger{}

	// Assume variables substitution on URL, Topic and Payload
	// Hack: the Path has been concatened with EscapedPath() (from net/url.go)
	// We must re-convert strings like $%7Bxyz%7D into ${xyz} to make variable substitution work !
	unescapedURL := RedecodeEscapedPath(mqttAction.URL)
	realURL := SubstParams(sessionMap, unescapedURL)

	connOpts := &MQTT.ClientOptions{
		ClientID:             mqttAction.ClientID,
		CleanSession:         true,
		AutoReconnect:        true,
		MaxReconnectInterval: mqttAction.MaxReconnectInterval * time.Second,
		KeepAlive:            int64(mqttAction.KeepAlive * time.Second),
		TLSConfig:            mqttAction.TLSConfig,
		Username:             mqttAction.Username,
		Password:             mqttAction.Password,
	}

	connOpts.AddBroker(realURL)
	log.Debugf("connOpts: %v", connOpts)

	mqttClient := MQTT.NewClient(connOpts)
	start := time.Now()
	token := mqttClient.Connect()
	token.Wait()
	//if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
	if err := token.Error(); err != nil {
		log.Error(err) // should report error in results
		return false
	}
	log.Debug("[MQTT] Connected")

	token = mqttClient.Publish(
		SubstParams(sessionMap, mqttAction.Topic),
		mqttAction.Qos,
		false,
		SubstParams(sessionMap, mqttAction.Payload))
	token.Wait()
	log.Debugf("[MQTT] Publish of: %s", mqttAction.Payload)
	mqttClient.Disconnect(0)
	elapsed := time.Since(start)

	/*
		quit := make(chan struct{})
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		go func() {
			<-c
			mqttClient.Disconnect(250)
			fmt.Println("[MQTT] Disconnected")

			quit <- struct{}{}
		}()
		<-quit
	*/

	sampleReqResult := buildSampleResult("MQTT", sessionMap["UID"], 0, 0, elapsed.Nanoseconds(), mqttAction.Title, realURL)
	resultsChannel <- sampleReqResult

	return true
}
