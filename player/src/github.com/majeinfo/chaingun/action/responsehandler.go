package action

import (
	"bytes"
	"errors"
	"github.com/JumboInteractiveLimited/jsonpath"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
	"gopkg.in/xmlpath.v2"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ResponseHandler describes the actions to apply on returned data
type ResponseHandler struct {
	FromHeader   string           `yaml:"from_header"`
	Jsonpaths    []*jsonpath.Path `yaml:"jsonpath"`
	Xmlpath      *xmlpath.Path    `yaml:"xmlpath"`
	Regex        *regexp.Regexp   `yaml:"regex"`
	Variable     string           `yaml:"variable"`
	Index        string           `yaml:"index"`
	Defaultvalue string           `yaml:"default_value"`
}

// NewResponseHandlers builds all the ResponseHandler from the Action described in YAML Playbook
func NewResponseHandlers(a map[interface{}]interface{}) ([]ResponseHandler, bool) {
	log.Debugf("NewResponseHandlers")

	valid := true
	var responseHandlers []ResponseHandler
	if a["responses"] == nil {
		responseHandlers = nil
	} else {
		switch v := a["responses"].(type) {
		case []interface{}:
			responseHandlers = make([]ResponseHandler, len(v))
			for idx, r1 := range v {
				r2 := r1.(map[interface{}]interface{})
				newResponse, err := NewResponseHandler(r2)
				if err != nil {
					valid = false
					break
				}
				//responseHandlers = append(responseHandlers, newResponse)
				responseHandlers[idx] = newResponse
			}
		default:
			log.Error("responses format is invalid")
			valid = false
		}
	}

	return responseHandlers, valid
}

// NewResponseHandler builds the ResponseHandler from the Action described in YAML playbook
func NewResponseHandler(a map[interface{}]interface{}) (ResponseHandler, error) {
	valid := true
	var responseHandler ResponseHandler

	validIndex := []string{"first", "last", "random"}
	if a["index"] != nil && !config.StringInSlice(a["index"].(string), validIndex) {
		log.Error("HttpAction ResponseHandler must define an Index of either of: first, last or random.")
		valid = false
	}
	if a["index"] == nil {
		a["index"] = "first"
	}
	if (a["jsonpath"] == nil || a["jsonpath"] == "") && (a["xmlpath"] == nil || a["xmlpath"] == "") && (a["regex"] == nil || a["regex"] == "") {
		log.Error("HttpAction ResponseHandler must define a Regexp, a Jsonpath or a Xmlpath.")
		valid = false
	}
	if (a["jsonpath"] != nil && a["jsonpath"] != "") && (a["xmlpath"] != nil && a["xmlpath"] != "") && (a["regex"] != nil && a["regex"] != "") {
		log.Error("HttpAction ResponseHandler can only define either a Regexp, a Jsonpath OR a Xmlpath.")
		valid = false
	}

	/*
		if !valid {
			log.Fatalf("Your YAML definition contains an invalid Action, see errors listed above.")
			valid = false
		}
	*/

	if a["jsonpath"] != nil && a["jsonpath"] != "" {
		var err error
		//responseHandler.Jsonpath = response["jsonpath"].(string)
		responseHandler.Jsonpaths, err = jsonpath.ParsePaths(a["jsonpath"].(string))
		if err != nil {
			log.Errorf("Jsonpath could not be compiled: %s", a["jsonpath"].(string))
			valid = false
		}
	}
	if a["xmlpath"] != nil && a["xmlpath"] != "" {
		// TODO perhaps compile Xmlpath expressions so we can validate early?
		//responseHandler.Xmlpath = response["xmlpath"].(string)
		var err error
		responseHandler.Xmlpath, err = xmlpath.Compile(a["xmlpath"].(string))
		if err != nil {
			log.Errorf("XmlPath could not be compiled: %s", a["xmlpath"].(string))
			valid = false
		}
	}
	if a["regex"] != nil && a["regex"] != "" {
		var err error
		responseHandler.Regex, err = regexp.Compile(a["regex"].(string))
		if err != nil {
			log.Errorf("Regexp could not be compiled: %s", a["regex"].(string))
			valid = false
		}
	}
	if a["default_value"] == nil {
		a["default_value"] = ""
	}

	if a["variable"] != nil {
		responseHandler.Variable = a["variable"].(string)
	}
	if a["index"] != nil {
		responseHandler.Index = a["index"].(string)
	}
	if a["default_value"] != nil {
		responseHandler.Defaultvalue = a["default_value"].(string)
	}
	if a["from_header"] != nil {
		responseHandler.FromHeader = a["from_header"].(string)
	}

	if !valid {
		return responseHandler, errors.New("Errors occurred during Response block analysis")
	}

	return responseHandler, nil
}

/**
 *  Extract data from response according to the desired processor
 */
func processResult(responseHandlers []ResponseHandler, sessionMap map[string]string, vulog *log.Entry, responseBody []byte, respHeader http.Header) bool {
	log.Debugf("processResult")
	for _, res := range responseHandlers {
		log.Debugf("responseHandlers")
		if !_processResult(res, sessionMap, vulog, responseBody, respHeader) {
			return false
		}
	}

	return true
}

func _processResult(responseHandler ResponseHandler, sessionMap map[string]string, vulog *log.Entry, responseBody []byte, respHeader http.Header) bool {
	if responseHandler.Jsonpaths != nil {
		return JSONProcessor(responseHandler, sessionMap, vulog, responseBody)
	}

	if responseHandler.Xmlpath != nil {
		return XMLPathProcessor(responseHandler, sessionMap, vulog, responseBody)
	}

	if responseHandler.Regex != nil {
		return RegexpProcessor(responseHandler, sessionMap, vulog, responseBody, respHeader)
	}

	return true
}

// JSONProcessor applies JSON expression to extract data from responses and fill variables
func JSONProcessor(responseHandler ResponseHandler, sessionMap map[string]string, vulog *log.Entry, responseBody []byte) bool {
	vulog.Debugf("Response processed by Json")

	eval, err := jsonpath.EvalPathsInBytes(responseBody, responseHandler.Jsonpaths)
	if err != nil {
		vulog.Errorf("Jsonpath failed to be applied: %v", err)
		return false
	}

	// TODO optimization: Don't reinitialize each time, reuse this somehow.
	max_values := 20
	resultsArray := make([]string, 0, max_values)
	idx := 0
	for {
		if result, ok := eval.Next(); ok {
			value := strings.TrimSpace(result.Pretty(false))
			vulog.Debugf("JSON extracted value: %s", value)
			idx++
			if idx > max_values {
				vulog.Errorf("Too many JSON values to extract (%d maximum), value %s ignored", max_values, value)
			} else {
				resultsArray = append(resultsArray, trimChar(value, '"'))
			}
		} else {
			break
		}
	}

	if eval.Error != nil {
		vulog.Errorf("Error while evaluating jsonpath: %s", eval.Error)
		return false
	}

	if len(resultsArray) == 0 {
		if responseHandler.Defaultvalue != "" {
			vulog.Warning("Jsonpath failed to apply, uses default value")
			resultsArray = append(resultsArray, responseHandler.Defaultvalue)
		} else {
			vulog.Errorf("Jsonpath %v failed to apply - no default value given", responseHandler.Jsonpaths)
			return false
		}
	}

	passResultIntoSessionMap(resultsArray, responseHandler, sessionMap, vulog)

	return true
}

// XMLPathProcessor extracts XML data from responses to fill variables
func XMLPathProcessor(responseHandler ResponseHandler, sessionMap map[string]string, vulog *log.Entry, responseBody []byte) bool {
	vulog.Debugf("Response processed by XmlPath")

	r := bytes.NewReader(responseBody)
	root, err := xmlpath.Parse(r)
	if err != nil {
		//log.Errorf("Could not parse reponse of page %s, as XML data: %s", httpAction.Title, err)
		vulog.Errorf("Could not parse reponse of page as XML data: %s", err)
		return false
	}

	iterator := responseHandler.Xmlpath.Iter(root)
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
		passResultIntoSessionMap(resultsArray, responseHandler, sessionMap, vulog)
	}

	return true
}

// RegexpProcessor applies the Regexp from responses to fill variables
func RegexpProcessor(responseHandler ResponseHandler, sessionMap map[string]string, vulog *log.Entry, responseBody []byte, respHeader http.Header) bool {
	var r string

	vulog.Debugf("Response processed by Regexp")

	// Two cases: extract value from Body or HTTP Header ?
	if responseHandler.FromHeader != "" {
		if respHeader.Get(responseHandler.FromHeader) != "" {
			r = respHeader.Get(responseHandler.FromHeader)
		} else {
			r = responseHandler.Defaultvalue
		}
		vulog.Debugf("Test matching against Header:")
	} else {
		r = string(responseBody[:])
		vulog.Debugf("Test matching against Body:")
	}
	log.Debug(r)

	res := responseHandler.Regex.FindAllStringSubmatch(r, -1)
	vulog.Debugf("Regex applied: %v", res)
	if res != nil && len(res) > 0 {
		vulog.Debugf("Regexp matches at least: %v, count of matching substring=%d", res[0], len(res))
		resultsArray := make([]string, len(res))
		for i := 0; i < len(res); i++ {
			//resultsArray = append(resultsArray, res[i][1])
			// If the rgexp did not capture anything (or lack parenthesis): nothing to store
			if len(res[i]) > 1 {
				resultsArray[i] = res[i][1]
			} else {
				resultsArray[i] = ""
				vulog.Debug("The regex matched but nothing captured !")
			}
		}
		passResultIntoSessionMap(resultsArray, responseHandler, sessionMap, vulog)
	} else {
		if responseHandler.Defaultvalue != "" {
			vulog.Warning("Regexp failed to apply, uses default value")
			resultsArray := make([]string, 1)
			resultsArray = append(resultsArray, responseHandler.Defaultvalue)
			passResultIntoSessionMap(resultsArray, responseHandler, sessionMap, vulog)
		} else {
			vulog.Errorf("Regexp '%s' failed to apply - no default value given", responseHandler.Regex)
			return false
		}
	}

	return true
}

func passResultIntoSessionMap(resultsArray []string, responseHandler ResponseHandler, sessionMap map[string]string, vulog *log.Entry) {
	vulog.Debugf("resultsArray=%v", resultsArray)

	if resultCount := len(resultsArray); resultCount > 0 {
		switch responseHandler.Index {
		case config.RE_FIRST:
			vulog.Debugf("First matching value: %s", resultsArray[0])
			sessionMap[responseHandler.Variable] = resultsArray[0]
			break
		case config.RE_LAST:
			vulog.Debugf("Last matching value: %s", resultsArray[resultCount-1])
			sessionMap[responseHandler.Variable] = resultsArray[resultCount-1]
			break
		case config.RE_RANDOM:
			if resultCount > 1 {
				sessionMap[responseHandler.Variable] = resultsArray[rand.Intn(resultCount)]
			} else {
				sessionMap[responseHandler.Variable] = resultsArray[0]
			}
			vulog.Debugf("Random matching value: %s", sessionMap[responseHandler.Variable])
			break
		default:
			vulog.Errorf("Internal error")
		}

	} else {
		// TODO how to handle requested, but missing result?
		vulog.Errorf("No value found in Response")
	}
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

func buildSampleResult(actionType string, vid string, contentLength int, status int, elapsed int64, title string, fullreq string) reporter.SampleReqResult {
	sampleReqResult := reporter.SampleReqResult{
		Vid:         vid,
		Type:        actionType,
		Latency:     elapsed,
		Size:        contentLength,
		Status:      status,
		Title:       title,
		When:        time.Since(reporter.SimulationStart).Nanoseconds(),
		FullRequest: fullreq,
	}
	return sampleReqResult
}
