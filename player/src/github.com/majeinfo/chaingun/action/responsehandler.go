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

type ResponseHandler struct {
	FromHeader   string           `yaml:"from_header"`
	Jsonpaths    []*jsonpath.Path `yaml:"jsonpath"`
	Xmlpath      *xmlpath.Path    `yaml:"xmlpath"`
	Regex        *regexp.Regexp   `yaml:"regex"`
	Variable     string           `yaml:"variable"`
	Index        string           `yaml:"index"`
	Defaultvalue string           `yaml:"default_value"`
}

// Build all the ResponseHandler from the Action described in YAML Playbook
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

// Build the ResponseHandler from the Action described in YAML playbook
func NewResponseHandler(a map[interface{}]interface{}) (ResponseHandler, error) {
	valid := true
	var responseHandler ResponseHandler

	valid_index := []string{"first", "last", "random"}
	if a["index"] != nil && !config.StringInSlice(a["index"].(string), valid_index) {
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
		return responseHandler, errors.New("Errors occurred during Response block analysis.")
	}

	return responseHandler, nil
}

/**
 *  Extract data from response according to the desired processor
 */
func processResult(responseHandlers []ResponseHandler, sessionMap map[string]string, responseBody []byte, respHeader http.Header) bool {
	log.Debugf("processResult")
	for _, res := range responseHandlers {
		log.Debugf("responseHandlers")
		if !_processResult(res, sessionMap, responseBody, respHeader) {
			return false
		}
	}

	return true
}

func _processResult(responseHandler ResponseHandler, sessionMap map[string]string, responseBody []byte, respHeader http.Header) bool {
	if responseHandler.Jsonpaths != nil {
		return JsonProcessor(responseHandler, sessionMap, responseBody)
	}

	if responseHandler.Xmlpath != nil {
		return XmlPathProcessor(responseHandler, sessionMap, responseBody)
	}

	if responseHandler.Regex != nil {
		return RegexpProcessor(responseHandler, sessionMap, responseBody, respHeader)
	}

	return true
}

func JsonProcessor(responseHandler ResponseHandler, sessionMap map[string]string, responseBody []byte) bool {
	log.Debugf("Response processed by Json")

	eval, err := jsonpath.EvalPathsInBytes(responseBody, responseHandler.Jsonpaths)
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
		if responseHandler.Defaultvalue != "" {
			log.Warning("Jsonpath failed to apply, uses default value")
			resultsArray = append(resultsArray, responseHandler.Defaultvalue)
		} else {
			log.Errorf("Jsonpath %s failed to apply - no default value given", responseHandler.Jsonpaths)
			return false
		}
	}

	passResultIntoSessionMap(resultsArray, responseHandler, sessionMap)

	return true
}

func XmlPathProcessor(responseHandler ResponseHandler, sessionMap map[string]string, responseBody []byte) bool {
	log.Debugf("Response processed by XmlPath")

	r := bytes.NewReader(responseBody)
	root, err := xmlpath.Parse(r)
	if err != nil {
		//log.Errorf("Could not parse reponse of page %s, as XML data: %s", httpAction.Title, err)
		log.Errorf("Could not parse reponse of page as XML data: %s", err)
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
		passResultIntoSessionMap(resultsArray, responseHandler, sessionMap)
	}

	return true
}

func RegexpProcessor(responseHandler ResponseHandler, sessionMap map[string]string, responseBody []byte, respHeader http.Header) bool {
	var r string = ""

	log.Debugf("Response processed by Regexp")

	// Two cases: extract value from Body or HTTP Header ?
	if responseHandler.FromHeader != "" {
		if respHeader.Get(responseHandler.FromHeader) != "" {
			r = respHeader.Get(responseHandler.FromHeader)
		} else {
			r = responseHandler.Defaultvalue
		}
	} else {
		r = string(responseBody[:])
	}

	res := responseHandler.Regex.FindAllStringSubmatch(r, -1)
	log.Debugf("Regex applied: %v", res)
	if res != nil && len(res) > 0 {
		log.Debugf("Regexp matches: %v", res[0])
		resultsArray := make([]string, 0, 10)
		if len(res[0]) > 1 {
			resultsArray = append(resultsArray, res[0][1])
		} else {
			resultsArray = append(resultsArray, res[0][0])
		}
		passResultIntoSessionMap(resultsArray, responseHandler, sessionMap)
	} else {
		if responseHandler.Defaultvalue != "" {
			log.Warning("Regexp failed to apply, uses default value")
			resultsArray := make([]string, 0, 10)
			resultsArray = append(resultsArray, responseHandler.Defaultvalue)
			passResultIntoSessionMap(resultsArray, responseHandler, sessionMap)
		} else {
			log.Errorf("Regexp '%s' failed to apply - no default value given", responseHandler.Regex)
			return false
		}
	}

	return true
}

func passResultIntoSessionMap(resultsArray []string, responseHandler ResponseHandler, sessionMap map[string]string) {
	resultCount := len(resultsArray)

	if resultCount > 0 {
		switch responseHandler.Index {
		case config.RE_FIRST:
			log.Debugf("First matching value: %s", resultsArray[0])
			sessionMap[responseHandler.Variable] = resultsArray[0]
			break
		case config.RE_LAST:
			log.Debugf("Last matching value: %s", resultsArray[resultCount-1])
			sessionMap[responseHandler.Variable] = resultsArray[resultCount-1]
			break
		case config.RE_RANDOM:
			if resultCount > 1 {
				sessionMap[responseHandler.Variable] = resultsArray[rand.Intn(resultCount-1)]
			} else {
				sessionMap[responseHandler.Variable] = resultsArray[0]
			}
			log.Debugf("Random matching value: %s", sessionMap[responseHandler.Variable])
			break
		default:
			log.Errorf("Internal error")
		}

	} else {
		// TODO how to handle requested, but missing result?
		log.Errorf("No value found in Response")
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

func buildSampleResult(action_type string, vid string, contentLength int, status int, elapsed int64, title string) reporter.SampleReqResult {
	sampleReqResult := reporter.SampleReqResult{
		vid,
		action_type,
		elapsed,
		contentLength,
		status,
		title,
		time.Since(reporter.SimulationStart).Nanoseconds(),
	}
	return sampleReqResult
}
