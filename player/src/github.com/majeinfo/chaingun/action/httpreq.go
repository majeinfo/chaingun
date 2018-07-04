package action

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/xmlpath.v2"
	"github.com/JumboInteractiveLimited/jsonpath"
	"bytes"
	"crypto/tls"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
)

var cookie_prefix = "__cookie__"
var cookie_prefix_length = len(cookie_prefix)

// Accepts a Httpaction and a one-way channel to write the results to.
func DoHttpRequest(httpAction HttpAction, resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string) bool {
	req := buildHttpRequest(httpAction, sessionMap)
	if req.Method != "POST" {
		log.Debugf("New Request: Method:%s, URL: %s", req.Method, req.URL)

	} else {
		log.Debugf("New Request: Method:%s, URL: %s, Body: %s", req.Method, req.URL, req.Body)
	}

	start := time.Now()
	var DefaultTransport http.RoundTripper = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	resp, err := DefaultTransport.RoundTrip(req)

	if err != nil {
		log.Errorf("HTTP request failed: %s", err)
		return false
	} 

	elapsed := time.Since(start)
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Fatal(err)
		log.Printf("Reading HTTP response failed: %s", err)
		httpReqResult := buildHttpResult(sessionMap["UID"], 0, resp.StatusCode, elapsed.Nanoseconds(), httpAction.Title)
		resultsChannel <- httpReqResult
		return false
	} 

	log.Debugf("Received data: %s", responseBody)
	defer resp.Body.Close()

	if httpAction.StoreCookie != "" {
		for _, cookie := range resp.Cookies() {
			if cookie.Name == httpAction.StoreCookie || httpAction.StoreCookie == "__all__" {
				log.Debugf("Store cookie: %s=%s", httpAction.StoreCookie, cookie.Value)
				sessionMap[cookie_prefix+cookie.Name] = cookie.Value
			}
		}
	}

	// if action specifies response action, parse using regexp/jsonpath
	if !processResult(httpAction, sessionMap, responseBody) {
		return false
	}
	httpReqResult := buildHttpResult(sessionMap["UID"], len(responseBody), resp.StatusCode, elapsed.Nanoseconds(), httpAction.Title)
	resultsChannel <- httpReqResult
	return true
}

func buildHttpResult(vid string, contentLength int, status int, elapsed int64, title string) reporter.HttpReqResult {
	httpReqResult := reporter.HttpReqResult{
		vid,
		"HTTP",
		elapsed,
		contentLength,
		status,
		title,
		time.Since(reporter.SimulationStart).Nanoseconds(),
	}
	return httpReqResult
}

func buildHttpRequest(httpAction HttpAction, sessionMap map[string]string) *http.Request {
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

	// Add headers
	if httpAction.Method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Add("Accept", httpAction.Accept)
	//req.Header.Add("Connection", "Keep-Alive")
	if httpAction.ContentType != "" {
		req.Header.Add("Content-Type", httpAction.ContentType)
	}

	// Add cookies stored by subsequent requests in the sessionMap having the kludgy __cookie__ prefix
	for key, value := range sessionMap {
		if strings.HasPrefix(key, cookie_prefix) {
			cookie := http.Cookie{
				Name:  key[cookie_prefix_length:len(key)],
				Value: value,
			}

			req.AddCookie(&cookie)
		}
	}

	return req
}

/**
 *  Extract data from response according to the desired processor
 */
func processResult(httpAction HttpAction, sessionMap map[string]string, responseBody []byte) bool {
	if httpAction.ResponseHandler.Jsonpaths != nil {
		return JsonProcessor(httpAction, sessionMap, responseBody)
	}

	if httpAction.ResponseHandler.Xmlpath != nil {
		return XmlPathProcessor(httpAction, sessionMap, responseBody)
	}

	if httpAction.ResponseHandler.Regex != nil {
		return RegexpProcessor(httpAction, sessionMap, responseBody)
	}

	return true
}

func JsonProcessor(httpAction HttpAction, sessionMap map[string]string, responseBody []byte) bool {
	log.Debugf("Response processed by Json")

	eval, err := jsonpath.EvalPathsInBytes(responseBody, httpAction.ResponseHandler.Jsonpaths)
	if err != nil {
		log.Errorf("Jsonpath failed to be applied: %v", err)
		return false
	}

	// TODO optimization: Don't reinitialize each time, reuse this somehow.
	resultsArray := make([]string, 0, 10)
	for {
		if result, ok := eval.Next(); ok {
			value := strings.TrimSpace(result.Pretty(false))
			log.Debugf("JSON extracted value: %s", value)
			resultsArray = append(resultsArray, trimChar(value, '"'))
		} else {
			break
		}
	}
	if eval.Error != nil {
		log.Errorf("Error while evaluating jsonpath: %s", eval.Error)
		return false
	}

	if len(resultsArray) == 0 {
		if httpAction.ResponseHandler.Defaultvalue != "" {
			log.Warning("Jsonpath failed to apply, uses default value")
			resultsArray = append(resultsArray, httpAction.ResponseHandler.Defaultvalue)		
		} else {
			log.Errorf("Jsonpath failed to apply - no default value given")
			return false
		}		
	}

	passResultIntoSessionMap(resultsArray, httpAction, sessionMap)

	return true
}

func XmlPathProcessor(httpAction HttpAction, sessionMap map[string]string, responseBody []byte) bool {
	log.Debugf("Response processed by XmlPath")

	r := bytes.NewReader(responseBody)
	root, err := xmlpath.Parse(r)
	if err != nil {
		log.Errorf("Could not parse reponse of page %s, as XML data: %s", httpAction.Title, err)
		return false
	}

	iterator := httpAction.ResponseHandler.Xmlpath.Iter(root)
	hasNext := iterator.Next()
	if hasNext {
		resultsArray := make([]string, 0, 10)
		for {
			if hasNext {
				node := iterator.Node()
				resultsArray = append(resultsArray, node.String())
				hasNext = iterator.Next()
			} else {
				break
			}
		}
		passResultIntoSessionMap(resultsArray, httpAction, sessionMap)
	}

	return true
}

func RegexpProcessor(httpAction HttpAction, sessionMap map[string]string, responseBody []byte) bool {
	log.Debugf("Response processed by Regexp")

	r := string(responseBody[:])
	res := httpAction.ResponseHandler.Regex.FindAllStringSubmatch(r, -1)
	log.Debugf("Regex applied: %v", res)
	if len(res) > 0 {
		// TODO: value should be computed like "abc$1$xyz" (config)
		resultsArray := make([]string, 0, 10)
		resultsArray = append(resultsArray, res[0][1])
		passResultIntoSessionMap(resultsArray, httpAction, sessionMap)
	} else {
		if httpAction.ResponseHandler.Defaultvalue != "" {
			log.Warning("Regexp failed to apply, uses default value")
			resultsArray := make([]string, 0, 10)
			resultsArray = append(resultsArray, httpAction.ResponseHandler.Defaultvalue)
			passResultIntoSessionMap(resultsArray, httpAction, sessionMap)			
		} else {
			log.Errorf("Regexp '%s' failed to apply", httpAction.ResponseHandler.Regex)
			return false
		}
	}

	return true
}

/**
 * Trims leading and trailing byte r from string s
 */
func trimChar(s string, r byte) string {
	sz := len(s)

	if sz > 0 && s[sz-1] == r {
		s = s[:sz-1]
	}
	sz = len(s)
	if sz > 0 && s[0] == r {
		s = s[1:sz]
	}
	return s
}

func passResultIntoSessionMap(resultsArray []string, httpAction HttpAction, sessionMap map[string]string) {
	resultCount := len(resultsArray)

	if resultCount > 0 {
		switch httpAction.ResponseHandler.Index {
		case config.RE_FIRST:
			sessionMap[httpAction.ResponseHandler.Variable] = resultsArray[0]
			break
		case config.RE_LAST:
			sessionMap[httpAction.ResponseHandler.Variable] = resultsArray[resultCount-1]
			break
		case config.RE_RANDOM:
			if resultCount > 1 {
				sessionMap[httpAction.ResponseHandler.Variable] = resultsArray[rand.Intn(resultCount-1)]
			} else {
				sessionMap[httpAction.ResponseHandler.Variable] = resultsArray[0]
			}
			break
		}

	} else {
		// TODO how to handle requested, but missing result?
	}
}
