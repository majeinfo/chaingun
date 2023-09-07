// TODO: a lot of go routines remain due to the elasticsearch interface.
//       should use a ConnectionPool ?
package action

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"

	log "github.com/sirupsen/logrus"
)

type ESClientContext struct {
	client *elasticsearch.Client
	ctx    context.Context
}

const (
	ES_ERR  = 500
	ES_JSON = 501
)

// DoESBRequest accepts a ESDBAction and a one-way channel to write the results to.
func DoESRequest(esAction ESAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	var trace_req string
	var client *elasticsearch.Client
	var ctx context.Context
	var err error
	var responseBody []byte

	index := SubstParams(sessionMap, esAction.Index, vulog)

	sampleReqResult := buildSampleResult(REPORTER_ES, sessionMap["UID"], 0, reporter.NETWORK_ERROR, 0, esAction.Title, "")

	if must_trace_request {
		trace_req = fmt.Sprintf("%s %s", esAction.Server, esAction.Command)
	} else {
		vulog.Debugf("New Request: URL: %s, Command: %s", esAction.Server, esAction.Command)
	}

	if !playbook.PersistentDBConn || vucontext.InitObject == nil { // persistent
		// Try to substitute the server name by an IP address
		server := esAction.Server
		if !disable_dns_cache {
			url, err := url.Parse(esAction.Server)
			if err != nil {
				if addr, status := utils.GetServerAddress(url.Host); status == true {
					url.Host = addr
					server = url.String()
				}
			}
		}

		vulog.Debugf("Create new ES Client")
		cfg := elasticsearch.Config{
			Addresses: []string{server},
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   10,
				ResponseHeaderTimeout: time.Duration(playbook.Timeout) * time.Second,
				DialContext:           (&net.Dialer{Timeout: time.Duration(playbook.Timeout) * time.Second}).DialContext,
				TLSClientConfig: &tls.Config{
					MinVersion:         tls.VersionTLS12,
					InsecureSkipVerify: true,
				},
			},
		}

		client, err = elasticsearch.NewClient(cfg)
		if err != nil {
			vulog.Errorf("Error creating the client: %s", err)
			completeSampleResult(&sampleReqResult, 0, reporter.NETWORK_ERROR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		} else {
			info, _ := client.Info()
			vulog.Debugf("%v", info)
		}

		clientContext := ESClientContext{client, ctx}
		vucontext.InitObject = &clientContext
	} else {
		vulog.Debugf("Reuse connection")
		clientContext := vucontext.InitObject.(*ESClientContext)
		client = clientContext.client
		ctx = clientContext.ctx
	}

	if !playbook.PersistentDBConn {
		//defer client.Close(context.TODO())
	} else {
		vucontext.CloseFunc = es_disconnect
	}

	var start time.Time = time.Now()
	//var response []byte

	switch esAction.Command {
	case ES_CREATEINDEX:
		settings := SubstParams(sessionMap, esAction.Settings, vulog)

                req := esapi.IndicesCreateRequest{
                        Index: index,
                        Body: strings.NewReader(settings),
                }

		ctx, _ := context.WithTimeout(context.Background(), time.Duration(playbook.Timeout)*time.Second)
		res, err := req.Do(ctx, client)
		if err != nil {
			vulog.Errorf("Error creating the Index: %s", err)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		defer res.Body.Close()
		vulog.Debugf("Response: %v", res)
		if res.IsError() {
			vulog.Errorf("Error creating the Index %s: %v", index, res)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, res.Status())
			resultsChannel <- sampleReqResult
			return false
		}

		vulog.Debugf("Index %s created", index)

	case ES_DELETEINDEX:
                req := esapi.IndicesDeleteRequest{
                        Index: []string{index},
                }

		ctx, _ := context.WithTimeout(context.Background(), time.Duration(playbook.Timeout)*time.Second)
		res, err := req.Do(ctx, client)
		if err != nil {
			vulog.Errorf("Error deleting the Index: %s", err)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		defer res.Body.Close()
		vulog.Debugf("Response: %v", res)
		if res.IsError() {
			vulog.Errorf("Error deleting the Index %s: %v", index, res)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, res.Status())
			resultsChannel <- sampleReqResult
			return false
		}

		vulog.Debugf("Index %s deleted", index)

	case ES_INSERT:
		document := SubstParams(sessionMap, esAction.Document, vulog)
		req := esapi.IndexRequest{
			Index: index,
			//DocumentID: strconv.Itoa(i + 1),
			Body: strings.NewReader(document),
			Refresh:    strconv.FormatBool(esAction.Refresh),
		}

		ctx, _ := context.WithTimeout(context.Background(), time.Duration(playbook.Timeout)*time.Second)
		res, err := req.Do(ctx, client)
		if err != nil {
			vulog.Errorf("ES insert action failed: %s", err)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		defer res.Body.Close()

		vulog.Debugf("Response: %v", res)
		if res.IsError() {
			vulog.Errorf("Error indexing document: %v", res)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, res.Status())
			resultsChannel <- sampleReqResult
			return false
		}

		vulog.Debugf("Insert done")

	case ES_SEARCH:
		query := SubstParams(sessionMap, esAction.Query, vulog)
		req := esapi.SearchRequest{
			Index: []string{index},
			Body: strings.NewReader(query),
		}

		ctx, _ := context.WithTimeout(context.Background(), time.Duration(playbook.Timeout)*time.Second)
		res, err := req.Do(ctx, client)

		//res, err := client.Search(
		//	client.Search.WithContext(ctx),
		//	client.Search.WithIndex(index),
		//	client.Search.WithBody(strings.NewReader(query)),
		//)
		if err != nil {
			vulog.Errorf("ES search action failed: %s", err)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		defer res.Body.Close()

		vulog.Debugf("Response: %v", res)
		if res.IsError() {
			vulog.Errorf("Error searching document: %v", res)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, res.Status())
			resultsChannel <- sampleReqResult
			return false
		}

		responseBody, err = ioutil.ReadAll(res.Body)
		if err != nil {
			vulog.Errorf("ES search action failed when reading Body: %s", err)
			completeSampleResult(&sampleReqResult, 0, ES_ERR, 0, err.Error())
			resultsChannel <- sampleReqResult
			return false
		}
		vulog.Debugf("Search done")
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
	if valid && len(responseBody) > 0 && !processResult(esAction.ResponseHandlers, sessionMap, vulog, responseBody, nil) {
		valid = false
	}

	completeSampleResult(&sampleReqResult, 0, statusCode, elapsed.Nanoseconds(), "")
	resultsChannel <- sampleReqResult
	return valid
}

func es_disconnect(vucontext *config.VUContext) {
	//clientContext := vucontext.InitObject.(*ESClientContext)
	//client := clientContext.client
	//client.Disconnect(context.TODO())
}
