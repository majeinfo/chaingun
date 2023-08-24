package action

import (
	"context"
	"crypto/tls"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"strings"
	"time"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

const (
	REPORTER_KAFKA string = "KAFKA"
	KAFKA_ERR             = 1
)

// DoMongoDBRequest accepts a MongoDBAction and a one-way channel to write the results to.
func DoKafkaRequest(kafkaAction KafkaAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	var trace_req string
	//var conn *kafka.Conn
	//var ctx context.Context
	//var err error
	sampleReqResult := buildSampleResult(REPORTER_KAFKA, sessionMap["UID"], 0, reporter.NETWORK_ERROR, 0, kafkaAction.Title, "")

	topic := SubstParams(sessionMap, kafkaAction.Topic, vulog)
	title := SubstParams(sessionMap, kafkaAction.Title, vulog)
	brokers := strings.Split(SubstParamsNoEscape(sessionMap, kafkaAction.Brokers, vulog), ",")

	if must_trace_request {
		trace_req = fmt.Sprintf("%s %s", kafkaAction.Brokers, title)
	} else {
		vulog.Debugf("New Request: URL: %s, Topic: %s", kafkaAction.Brokers, topic)
	}

	// Note: persistent connection are not handled here
	// Note: the DNS cache is not handled neither
	// Note: implement SASL Mechanism
	// TODO: prise en compte des variables
	// TODO: prise en compte des résulats
	// TODO: possibilité de faire plusieurs read ?
	vulog.Debugf("Create new Kafka Client")
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(playbook.Timeout)*time.Second)
	var start time.Time = time.Now()

	switch kafkaAction.Command {
	case KAFKA_WRITE:
		w := &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.RoundRobin{},
			WriteTimeout: time.Duration(playbook.Timeout) * time.Second,
		}
		if kafkaAction.TLSEnabled {
			w.Transport = &kafka.Transport{
				TLS: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
		}
		defer w.Close()

		msgs := []kafka.Message{
			{
				Key:   []byte(kafkaAction.Key),
				Value: []byte(kafkaAction.Value),
			},
		}
		if err := w.WriteMessages(ctx, msgs...); err != nil {
			vulog.Errorf("Kafka write request failed: %s", err)
			buildKafkaSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		vulog.Debugf("Kafka action done (%d message(s) written)", len(msgs))

	case KAFKA_READ:
		dialer := &kafka.Dialer{
			Timeout: time.Duration(playbook.Timeout) * time.Second,
		}
		if kafkaAction.TLSEnabled {
			dialer.TLS = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			Dialer:  dialer,
		})
		defer r.Close()

		for {
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				vulog.Errorf("Kafka read request failed: %s", err)
				buildKafkaSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
				resultsChannel <- sampleReqResult
				return false
			}
			vulog.Debugf("msg read: %s", string(msg.Value))
			break // only one msg read
		}

	case KAFKA_CREATETOPIC:
		conn, err := kafka.Dial("tcp", brokers[0])
		if err != nil {
			vulog.Errorf("Kafka dial failed: %s", err)
			buildKafkaSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		defer conn.Close()

		topicConfigs := []kafka.TopicConfig{
			{
				Topic:             topic,
				NumPartitions:     1,
				ReplicationFactor: 1,
			},
		}

		err = conn.CreateTopics(topicConfigs...)
		if err != nil {
			vulog.Errorf("Kafka create topic failed: %s", err)
			buildKafkaSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		vulog.Debugf("Topic %s created", topic)

	case KAFKA_DELETETOPIC:
		conn, err := kafka.Dial("tcp", brokers[0])
		if err != nil {
			vulog.Errorf("Kafka dial failed: %s", err)
			buildKafkaSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		defer conn.Close()

		err = conn.DeleteTopics(topic)
		if err != nil {
			vulog.Errorf("Kafka delete topic failed: %s", err)
			buildKafkaSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		vulog.Debugf("Topic %s deleted", topic)
	}

	elapsed := time.Since(start)
	statusCode := 0

	if must_trace_request {
		vulog.Infof("%s", trace_req)
	}
	if must_display_srv_resp {
		vulog.Debugf("")
	}

	buildKafkaSampleResult(&sampleReqResult, 0, statusCode, elapsed.Nanoseconds(), "")
	resultsChannel <- sampleReqResult
	return true
}

func buildKafkaSampleResult(sample *reporter.SampleReqResult, contentLength int, status int, elapsed int64, fullreq string) {
	sample.Status = status
	sample.Size = contentLength
	sample.Latency = elapsed
	sample.FullRequest = fullreq
}

func kafka_disconnect(vucontext *config.VUContext) {
	//clientContext := vucontext.InitObject.(*MongoClientContext)
	//client := clientContext.client
	//client.Disconnect(context.TODO())

	conn := vucontext.InitObject.(*kafka.Conn)
	conn.Close()
}
