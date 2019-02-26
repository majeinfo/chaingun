package action

import (
	"github.com/gorilla/websocket"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
	"time"
)

// TODO: manage cookies !
func DoWSRequest(wsAction WSAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	//req := buildWSRequest(wsAction, sessionMap)
	log.Debugf("New Request: URL: %s", wsAction.Url)

	start := time.Now()

	// TODO: should be done once per VU and per script or this could be configurable (for HTTP/S too)
	c, _, err := websocket.DefaultDialer.Dial(wsAction.Url, nil)
	if err != nil {
		log.Errorf("WS Connection failed: %s", err)
		return false
	}
	defer c.Close()

	log.Debugf("wsAction.Body=%s", wsAction.Body)
	if wsAction.Body != "" {
		err = c.WriteMessage(websocket.TextMessage, []byte(wsAction.Body))
		if err != nil {
			log.Errorf("WS WriteMessage failed: %s", err)
			return false
		}
	}

	bodyLen := 0
	respCode := -1
	ok := true
	//if wsAction.ResponseHandler.Regex != nil || wsAction.ResponseHandler.Jsonpaths != nil || wsAction.ResponseHandler.Xmlpath != nil {
	if len(wsAction.ResponseHandlers) > 0 {
		log.Debugf("Starts a new timeout for %d seconds before reading the message", playbook.Timeout)
		ticker := time.NewTicker(time.Duration(playbook.Timeout) * time.Second)
		defer ticker.Stop()

		done := make(chan struct{})

		go func() {
			defer close(done)

			_, responseBody, err := c.ReadMessage()
			if err != nil {
				log.Errorf("WS ReadMessage error: %s", err)
				ok = false
			}
			log.Debugf("WS ReadMessage recv: %s", responseBody)

			// if action specifies response action, parse using regexp/jsonpath
			if !processResult(wsAction.ResponseHandlers, sessionMap, responseBody, nil) {
				ok = false
			} else {
				bodyLen = len(responseBody)
				respCode = 0
			}
		}()

		select {
		case <-done:
			break
		case t := <-ticker.C:
			log.Errorf("WS ReadMessage timeout", t)
			ok = false
			break
		}
	}

	elapsed := time.Since(start)
	log.Debugf("elapsed time=%d", elapsed)

	sampleReqResult := buildSampleResult("WS", sessionMap["UID"], bodyLen, respCode, elapsed.Nanoseconds(), wsAction.Title)
	resultsChannel <- sampleReqResult

	return ok
}

/*
func buildWSRequest(wsAction WSAction, sessionMap map[string]string) { //*http.Request {
	/*
	var req *http.Request
	var err error
	if httpAction.Body != "" {
		reader := strings.NewReader(SubstParams(sessionMap, httpAction.Body))
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, httpAction.Url), reader)
	} else if httpAction.Template != "" {
		reader := strings.NewReader(SubstParams(sessionMap, httpAction.Template))
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, httpAction.Url), reader)
	} else {
		req, err = http.NewRequest(httpAction.Method, SubstParams(sessionMap, httpAction.Url), nil)
	}
	if err != nil {
		log.Fatal(err)
	}

	return req
}
*/
