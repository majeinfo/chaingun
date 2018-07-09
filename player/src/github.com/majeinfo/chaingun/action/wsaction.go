package action

import (
    log "github.com/sirupsen/logrus"
    "github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/config"    
)

type WSAction struct {
    //Method          string              `yaml:"method"`
    Url             string              `yaml:"url"`
    Body            string              `yaml:"body"`
    Title           string              `yaml:"title"`
    StoreCookie     string              `yaml:"storeCookie"`
    ResponseHandlers []ResponseHandler 	`yaml:"responses"`
}

func (h WSAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
    return DoWSRequest(h, resultsChannel, sessionMap, playbook)
}

func NewWSAction(a map[interface{}]interface{}) WSAction {
    valid := true
    if a["url"] == "" || a["url"] == nil {
        log.Error("Error: WSAction must define a URL.")
        valid = false
    }
    if a["title"] == nil || a["title"] == "" {
        log.Error("Error: WSAction must define a title.")
        valid = false
    }

    var storeCookie string
    if a["storeCookie"] != nil && a["storeCookie"].(string) != "" {
        storeCookie = a["storeCookie"].(string)
    }

    var responseHandlers []ResponseHandler
    if a["responses"] == nil {
        responseHandlers = nil
    } else {
        switch v := a["responses"].(type) {
        case []interface {}:
            responseHandlers = make([]ResponseHandler, len(v))
            for _, r1 := range v {
                r2 := r1.(map[interface{}]interface{})
                newResponse, err := NewResponseHandler(r2)
                if err != nil {
                    valid = false
                    break
                }
                responseHandlers = append(responseHandlers, newResponse)
            }
        default:
            log.Error("responses format is invalid")
            valid = false
        }
    }

    if !valid {
        log.Fatalf("Your YAML Playbook contains an invalid WSAction, see errors listed above.")
    }


    WSAction := WSAction{
        a["url"].(string),
        getBody(a),
        a["title"].(string),
        storeCookie,
        responseHandlers,
    }

	log.Debugf("WSAction: %v", WSAction)
	
    return WSAction
}
