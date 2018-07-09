package action

import (
    log "github.com/sirupsen/logrus"
    "github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/config"    
)

type HttpAction struct {
    Method          string              `yaml:"method"`
    Url             string              `yaml:"url"`
    Body            string              `yaml:"body"`
    Template        string              `yaml:"template"`
    Accept          string              `yaml:"accept"`
    ContentType     string              `yaml:"contentType"`
    Title           string              `yaml:"title"`
    StoreCookie     string              `yaml:"storeCookie"`
    ResponseHandlers []ResponseHandler   `yaml:"responses"`
}

func (h HttpAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
    return DoHttpRequest(h, resultsChannel, sessionMap, playbook)
}

func NewHttpAction(a map[interface{}]interface{}) HttpAction {
    valid := true
    if a["url"] == "" || a["url"] == nil {
        log.Error("Error: HttpAction must define a URL.")
        valid = false
    }
    valid_methods := []string{"GET", "POST", "PUT", "HEAD", "DELETE"}
    if !stringInSlice(a["method"].(string), valid_methods) {
        log.Error("Error: HttpAction must specify a valid HTTP method: GET, POST, PUT, HEAD or DELETE")
        valid = false
    }
    if a["title"] == nil || a["title"] == "" {
        log.Error("Error: HttpAction must define a title.")
        valid = false
    }

    if a["body"] != nil && a["template"] != nil {
        log.Error("Error: A HttpAction can not define both a 'body' and a 'template'.")
        valid = false
    }

    accept := "text/html,application/json,application/xhtml+xml,application/xml,text/plain"
    if a["accept"] != nil && len(a["accept"].(string)) > 0 {
        accept = a["accept"].(string)
    }

    var contentType string
    if a["contentType"] != nil && len(a["contentType"].(string)) > 0 {
        contentType = a["contentType"].(string)
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
        log.Fatalf("Your YAML Playbook contains an invalid HTTPAction, see errors listed above.")
    }

    httpAction := HttpAction{
        a["method"].(string),
        a["url"].(string),
        getBody(a),
        getTemplate(a),
        accept,
        contentType,
        a["title"].(string),
        storeCookie,
        responseHandlers,
    }

    log.Debugf("HTTPAction: %v", httpAction)

    return httpAction
}
