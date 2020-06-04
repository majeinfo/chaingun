package action

import (
	"github.com/gorilla/websocket"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
	"time"
)

const (
	REPORTER_WS string = "WS"
)

// DoWSRequest handles requests made using WebSocket Protocol
// TODO: manage cookies !
func DoWSRequest(wsAction WSAction, resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	vulog.Debugf("New Request: URL: %s", wsAction.URL)

	start := time.Now()

	// Hack: the Path has been concatened with EscapedPath() (from net/url.go)
	// We must re-convert strings like $%7Bxyz%7D into ${xyz} to make variable substitution work !
	unescapedURL := RedecodeEscapedPath(wsAction.URL)
	url := SubstParams(sessionMap, unescapedURL, vulog)
	vulog.Debugf("Translated Request: URL: %s", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		vulog.Errorf("WS Connection failed: %s", err)
		return false
	}
	defer c.Close()

	vulog.Debugf("wsAction.Body=%s", wsAction.Body)
	if wsAction.Body != "" {
		body := SubstParams(sessionMap, wsAction.Body, vulog)
		err = c.WriteMessage(websocket.TextMessage, []byte(body))
		if err != nil {
			vulog.Errorf("WS WriteMessage failed: %s", err)
			return false
		}
	}

	bodyLen := 0
	respCode := -1
	ok := true

	if len(wsAction.ResponseHandlers) > 0 {
		vulog.Debugf("Starts a new timeout for %d seconds before reading the message", playbook.Timeout)
		ticker := time.NewTicker(time.Duration(playbook.Timeout) * time.Second)
		defer ticker.Stop()

		done := make(chan struct{})

		go func() {
			defer close(done)

			_, responseBody, err := c.ReadMessage()
			if err != nil {
				vulog.Errorf("WS ReadMessage error: %s", err)
				ok = false
			}
			vulog.Debugf("WS ReadMessage recv: %s", responseBody)

			// if action specifies response action, parse using regexp/jsonpath
			if !processResult(wsAction.ResponseHandlers, sessionMap, vulog, responseBody, nil) {
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
			vulog.Errorf("WS ReadMessage timeout %v", t)
			ok = false
			break
		}
	}

	elapsed := time.Since(start)
	vulog.Debugf("elapsed time=%d", elapsed)

	sampleReqResult := buildSampleResult(REPORTER_WS, sessionMap["UID"], bodyLen, respCode, elapsed.Nanoseconds(), wsAction.Title, wsAction.URL)
	resultsChannel <- sampleReqResult

	return ok
}
