package action

import (
	"errors"
	"regexp"
	"math/rand"
	"bytes"
	"strings"
	"time"
    log "github.com/sirupsen/logrus"
    "gopkg.in/xmlpath.v2"
    "github.com/JumboInteractiveLimited/jsonpath"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
)

type ResponseHandler struct {
    Jsonpaths []*jsonpath.Path `yaml:"jsonpath"`
    Xmlpath *xmlpath.Path `yaml:"xmlpath"`
    Regex *regexp.Regexp `yaml:"regex"`
    Variable string `yaml:"variable"`
    Index    string `yaml:"index"`
    Defaultvalue string `yaml:"default_value"`
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

// Build the ResponseHandler from the Action described in YAML playbook
func NewResponseHandler(a map[interface{}]interface{}) (ResponseHandler, error) {
	valid := true
	var responseHandler ResponseHandler
	
    if a["response"] != nil {
		r := a["response"].(map[interface{}]interface{})
		
        valid_index := []string{"first", "last", "random"}
        if r["index"] != nil && !stringInSlice(r["index"].(string), valid_index) {
            log.Error("Error: HttpAction ResponseHandler must define an Index of either of: first, last or random.")
            valid = false
        }
        if (r["jsonpath"] == nil || r["jsonpath"] == "") && (r["xmlpath"] == nil || r["xmlpath"] == "") && (r["regex"] == nil || r["regex"] == "") {
            log.Error("Error: HttpAction ResponseHandler must define a Regexp, a Jsonpath or a Xmlpath.")
            valid = false
        }
        if (r["jsonpath"] != nil && r["jsonpath"] != "") && (r["xmlpath"] != nil && r["xmlpath"] != "") && (r["regex"] != nil && r["regex"] != "") {
            log.Error("Error: HttpAction ResponseHandler can only define either a Regexp, a Jsonpath OR a Xmlpath.")
            valid = false
        }

		/*
		if !valid {
			log.Fatalf("Your YAML definition contains an invalid Action, see errors listed above.")
			valid = false
		}
		*/

        if r["jsonpath"] != nil && r["jsonpath"] != "" {
            var err error
            //responseHandler.Jsonpath = response["jsonpath"].(string)
            responseHandler.Jsonpaths, err = jsonpath.ParsePaths(r["jsonpath"].(string))
            if err != nil {
				log.Error("Jsonpath could not be compiled: %s", r["jsonpath"].(string))
				valid = false
            }
        }
        if r["xmlpath"] != nil && r["xmlpath"] != "" {
            // TODO perhaps compile Xmlpath expressions so we can validate early?            
            //responseHandler.Xmlpath = response["xmlpath"].(string)
            var err error
            responseHandler.Xmlpath, err = xmlpath.Compile(r["xmlpath"].(string))
            if err != nil {
				log.Error("XmlPath could not be compiled: %s", r["xmlpath"].(string))
				valid = false
			}
        }
        if r["regex"] != nil && r["regex"] != "" {
			var err error
            responseHandler.Regex, err = regexp.Compile(r["regex"].(string))
            if err != nil {
				log.Error("Regexp could not be compiled: %s", r["regex"].(string))
				valid = false
			}
        }
        if r["default_value"] == nil {
            r["default_value"] = ""
        }

        if r["variable"] != nil {
			responseHandler.Variable = r["variable"].(string)
		}
        if r["index"] != nil {
			responseHandler.Index = r["index"].(string)
		}
		if r["default_value"] != nil {
			responseHandler.Defaultvalue = r["default_value"].(string)
		}
	}
	
	if !valid {
		return responseHandler, errors.New("Errors occurred during Response block analysis.")
	}

    return responseHandler, nil
}

/**
 *  Extract data from response according to the desired processor
 */
 func processResult(responseHandler ResponseHandler, sessionMap map[string]string, responseBody []byte) bool {
	if responseHandler.Jsonpaths != nil {
		return JsonProcessor(responseHandler, sessionMap, responseBody)
	}

	if responseHandler.Xmlpath != nil {
		return XmlPathProcessor(responseHandler, sessionMap, responseBody)
	}

	if responseHandler.Regex != nil {
		return RegexpProcessor(responseHandler, sessionMap, responseBody)
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
			log.Errorf("Jsonpath failed to apply - no default value given")
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

// TODO: add flag to make regexp case-insensitive
func RegexpProcessor(responseHandler ResponseHandler, sessionMap map[string]string, responseBody []byte) bool {
	log.Debugf("Response processed by Regexp")

	r := string(responseBody[:])
	res := responseHandler.Regex.FindAllStringSubmatch(r, -1)
	log.Debugf("Regex applied: %v", res)
	if len(res) > 0 {
		// TODO: value should be computed like "abc$1$xyz" (config)
		log.Debugf("Regexp matches: %v", res[0])
		resultsArray := make([]string, 0, 10)
		resultsArray = append(resultsArray, res[0][0])
		passResultIntoSessionMap(resultsArray, responseHandler, sessionMap)
	} else {
		if responseHandler.Defaultvalue != "" {
			log.Warning("Regexp failed to apply, uses default value")
			resultsArray := make([]string, 0, 10)
			resultsArray = append(resultsArray, responseHandler.Defaultvalue)
			passResultIntoSessionMap(resultsArray, responseHandler, sessionMap)			
		} else {
			log.Errorf("Regexp '%s' failed to apply - no default value", responseHandler.Regex)
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
			sessionMap[responseHandler.Variable] = resultsArray[0]
			break
		case config.RE_LAST:
			sessionMap[responseHandler.Variable] = resultsArray[resultCount-1]
			break
		case config.RE_RANDOM:
			if resultCount > 1 {
				sessionMap[responseHandler.Variable] = resultsArray[rand.Intn(resultCount-1)]
			} else {
				sessionMap[responseHandler.Variable] = resultsArray[0]
			}
			break
		}

	} else {
		// TODO how to handle requested, but missing result?
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

