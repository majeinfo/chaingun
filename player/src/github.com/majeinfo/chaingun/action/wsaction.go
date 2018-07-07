package action

import (
    "os"
    log "github.com/sirupsen/logrus"
    "github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/config"    
)

type WSAction struct {
    //Method          string              `yaml:"method"`
    Url             string              `yaml:"url"`
    Body            string              `yaml:"body"`
    Title           string              `yaml:"title"`
    ResponseHandler ResponseHandler 	`yaml:"response"`
}

func (h WSAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
    return DoWSRequest(h, resultsChannel, sessionMap, playbook)
}

func NewWSAction(a map[interface{}]interface{}) WSAction {
    var valid bool = true
    if a["url"] == "" || a["url"] == nil {
        log.Error("Error: WSAction must define a URL.")
        valid = false
    }
    if a["title"] == nil || a["title"] == "" {
        log.Error("Error: WSAction must define a title.")
        valid = false
    }

    if a["response"] != nil {
        r := a["response"].(map[interface{}]interface{})
        valid_index := []string{"first", "last", "random"}
        if r["index"] != nil && !stringInSlice(r["index"].(string), valid_index) {
            log.Error("Error: WSAction ResponseHandler must define an Index of either of: first, last or random.")
            valid = false
        }
        if (r["jsonpath"] == nil || r["jsonpath"] == "") && (r["xmlpath"] == nil || r["xmlpath"] == "") && (r["regex"] == nil || r["regex"] == "") {
            log.Error("Error: WSAction ResponseHandler must define a Regexp, a Jsonpath or a Xmlpath.")
            valid = false
        }
        if (r["jsonpath"] != nil && r["jsonpath"] != "") && (r["xmlpath"] != nil && r["xmlpath"] != "") && (r["regex"] != nil && r["regex"] != "") {
            log.Error("Error: WSAction ResponseHandler can only define either a Regexp, a Jsonpath OR a Xmlpath.")
            valid = false
		}
		/*
        if r["variable"] == nil || r["variable"] == "" {
            log.Error("Error: WSAction ResponseHandler must define a Variable.")
            valid = false
		}
		*/
    }

    /*
    if !valid {
        log.Fatalf("Your YAML defintion contains an invalid WSAction, see errors listed above.")
    }
    */
    
    responseHandler, err := NewResponseHandler(a)
    if !valid || err != nil {
        os.Exit(1)
    }

    WSAction := WSAction{
        a["url"].(string),
        getBody(a),
        a["title"].(string),
        responseHandler,
    }

	log.Debugf("WSAction: %v", WSAction)
	
    return WSAction
}
