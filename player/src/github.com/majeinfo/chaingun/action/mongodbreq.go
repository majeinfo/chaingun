package action

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"

	log "github.com/sirupsen/logrus"
)

const (
	REPORTER_MONGODB string = "MONGODB"
	MONGODB_ERR             = 500
	MONGODB_JSON            = 501
)

// DoMongoDBRequest accepts a MongoDBAction and a one-way channel to write the results to.
func DoMongoDBRequest(mongodbAction MongoDBAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	var trace_req string

	if must_trace_request {
		trace_req = fmt.Sprintf("%s %s", mongodbAction.Server, mongodbAction.Command)
	} else {
		vulog.Debugf("New Request: URL: %s, Command: %s", mongodbAction.Server, mongodbAction.Command)
	}

	// Try to substitute the server name by an IP address
	server := mongodbAction.Server
	if !disable_dns_cache {
		url, err := url.Parse(mongodbAction.Server)
		if err != nil {
			if addr, status := utils.GetServerAddress(url.Host); status == true {
				url.Host = addr
				server = url.String()
			}
		}
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(server))
	ctx, _ := context.WithTimeout(context.Background(), time.Duration(playbook.Timeout)*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		vulog.Errorf("MongoDB request failed: %s", err)
		sampleReqResult := buildSampleResult(REPORTER_MONGODB, sessionMap["UID"], 0, reporter.NETWORK_ERROR, 0, mongodbAction.Title, err.Error())
		resultsChannel <- sampleReqResult
		return false
	}

	defer client.Disconnect(context.TODO())

	/*
		err = client.Ping(ctx, nil)
		if err != nil {
				log.Fatal(err)
		}
	*/

	collection := client.Database(mongodbAction.Database).Collection(mongodbAction.Collection)
	var bdoc interface{}
	var start time.Time = time.Now()
	var response []byte

	switch mongodbAction.Command {
	case "drop":
		err = collection.Drop(ctx)
		if err != nil {
			vulog.Errorf("MongoDB drop action failed: %s", err)
			sampleReqResult := buildSampleResult(REPORTER_MONGODB, sessionMap["UID"], 0, MONGODB_JSON, 0, mongodbAction.Title, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		vulog.Debugf("Drop collection done")

	case "insertone":
		doc := SubstParams(sessionMap, mongodbAction.Document, vulog)
		err = bson.UnmarshalExtJSON([]byte(doc), true, &bdoc)
		if err != nil {
			vulog.Errorf("MongoDB insertone action failed: %s", err)
			sampleReqResult := buildSampleResult(REPORTER_MONGODB, sessionMap["UID"], 0, MONGODB_JSON, 0, mongodbAction.Title, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}

		res, err := collection.InsertOne(ctx, &bdoc)
		if err != nil {
			vulog.Errorf("MongoDB insertone failed: %s", err)
			sampleReqResult := buildSampleResult(REPORTER_MONGODB, sessionMap["UID"], 0, MONGODB_ERR, 0, mongodbAction.Title, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		sessionMap[config.MONGODB_LAST_INSERT_ID] = res.InsertedID.(primitive.ObjectID).String() // ...but the string is not useful it should keep its original type !
		vulog.Debugf("Insert result: %v, ID=%v", res, res.InsertedID)

	case "findone":
		doc := SubstParams(sessionMap, mongodbAction.Filter, vulog)
		err = bson.UnmarshalExtJSON([]byte(doc), true, &bdoc)
		if err != nil {
			vulog.Errorf("MongoDB findone action failed: %s", err)
			sampleReqResult := buildSampleResult(REPORTER_MONGODB, sessionMap["UID"], 0, MONGODB_JSON, 0, mongodbAction.Title, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}

		find_res := collection.FindOne(ctx, &bdoc)
		err = find_res.Decode(&bdoc)
		if err != nil {
			vulog.Errorf("MongoDB findone failed: %s", err)
			sampleReqResult := buildSampleResult(REPORTER_MONGODB, sessionMap["UID"], 0, MONGODB_ERR, 0, mongodbAction.Title, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}

		response, err = bson.MarshalExtJSON(bdoc, true, false)
		if err != nil {
			vulog.Errorf("MongoDB findone marshal failed: %s", err)
			sampleReqResult := buildSampleResult(REPORTER_MONGODB, sessionMap["UID"], 0, MONGODB_JSON, 0, mongodbAction.Title, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		vulog.Debugf("FindOne gets: %v", response)
	}

	elapsed := time.Since(start)
	statusCode := 0

	if must_trace_request {
		vulog.Infof("%s", trace_req)
	}
	if must_display_srv_resp {
		vulog.Debugf("")
	}

	valid := true

	// if action specifies response action, parse using regexp/jsonpath
	if valid && len(response) > 0 && !processResult(mongodbAction.ResponseHandlers, sessionMap, vulog, response, nil) {
		valid = false
	}

	sampleReqResult := buildSampleResult(REPORTER_MONGODB, sessionMap["UID"], 0, statusCode, elapsed.Nanoseconds(), mongodbAction.Title, "")
	resultsChannel <- sampleReqResult
	return valid
}