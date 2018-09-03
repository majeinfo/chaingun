package action

import (
    "strings"

    log "github.com/sirupsen/logrus"
    "github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/config"    
)

type HttpAction struct {
    Method          string              `yaml:"method"`
    Url             string              `yaml:"url"`
    Body            string              `yaml:"body"`
    Template        string              `yaml:"template"`
    Headers         map[string]string   `yaml:"headers"`
    Title           string              `yaml:"title"`
    StoreCookie     string              `yaml:"storeCookie"`
    ResponseHandlers []ResponseHandler   `yaml:"responses"`
}

func (h HttpAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
    return DoHttpRequest(h, resultsChannel, sessionMap, playbook)
}

func NewHttpAction(a map[interface{}]interface{}, dflt config.Default) HttpAction {
    log.Debugf("NewhttpAction=%v", a)
    valid := true

    if a["url"] == "" || a["url"] == nil {
        log.Error("HttpAction must define a URL.")
        valid = false
    } else {
        valid = setDefaultURL(a, dflt)
    }

    if a["method"] == nil || a["method"] == "" {
        if dflt.Method == "" {
            log.Error("Action has no Method and no default Method specified")
            valid = false
        } else {
            a["method"] = dflt.Method
        }
    } else if !config.IsValidHTTPMethod(a["method"].(string)) {
        log.Error("HttpAction must specify a valid HTTP method: GET, POST, PUT, HEAD or DELETE")
        valid = false
    }
    if a["title"] == nil || a["title"] == "" {
        log.Error("HttpAction must define a title.")
        valid = false
    }

    if a["body"] != nil && a["template"] != nil {
        log.Error("A HttpAction can not define both a 'body' and a 'template'.")
        valid = false
    }

    var storeCookie string
    if a["storeCookie"] != nil && a["storeCookie"].(string) != "" {
        storeCookie = a["storeCookie"].(string)
    }

    headers := make(map[string]string, 20)
    if a["headers"] != nil {
        for hdr, value := range (a["headers"].(map[interface{}]interface{})) {
            log.Debugf("Header Key=%s / Value=%s", hdr.(string), value.(string))
            headers[strings.ToLower(hdr.(string))] = value.(string)
        }
    }
    if _, ok := headers["accept"]; !ok {
        headers["accept"] = "text/html,application/json,application/xhtml+xml,application/xml,text/plain"
    }
    headers["user-agent"] = "chaingun-by-JD"

    responseHandlers, valid_resp  := NewResponseHandlers(a)

    if !valid || !valid_resp {
        log.Fatalf("Your YAML Playbook contains an invalid HTTPAction, see errors listed above.")
    }

    httpAction := HttpAction{
        a["method"].(string),
        a["url"].(string),
        getBody(a),
        getTemplate(a),
        headers,
        a["title"].(string),
        storeCookie,
        responseHandlers,
    }

    log.Debugf("HTTPAction: %v", httpAction)

    return httpAction
}
