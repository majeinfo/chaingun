package action

import (
    "github.com/majeinfo/chaingun/config"
    "github.com/majeinfo/chaingun/reporter"
    log "github.com/sirupsen/logrus"
)

// WSAction describes the structure of WebSocket Action
type WSAction struct {
    //Method          string              `yaml:"method"`
    URL              string            `yaml:"url"`
    Body             string            `yaml:"body"`
    Title            string            `yaml:"title"`
    StoreCookie      string            `yaml:"storeCookie"`
    ResponseHandlers []ResponseHandler `yaml:"responses"`
}

// Execute a WebSocket Action
func (h WSAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
    return DoWSRequest(h, resultsChannel, sessionMap, vulog, playbook)
}

// NewWSAction builds a new WebSocket Action
func NewWSAction(a map[interface{}]interface{}, dflt config.Default) (WSAction, bool) {
    valid := true
    if a["url"] == "" || a["url"] == nil {
        log.Error("WSAction must define a URL.")
        a["url"] = ""
        valid = false
    } else {
        valid = setDefaultURL(a, dflt)
    }

    if a["title"] == nil || a["title"] == "" {
        log.Error("WSAction must define a title.")
        a["title"] = ""
        valid = false
    }

    var storeCookie string
    if a["storeCookie"] != nil && a["storeCookie"].(string) != "" {
        storeCookie = a["storeCookie"].(string)
    }

    responseHandlers, validResp := NewResponseHandlers(a)

    if !valid || !validResp {
        log.Errorf("Your YAML Playbook contains an invalid WSAction, see errors listed above.")
        valid = false
    }

    WSAction := WSAction{
        a["url"].(string),
        getBody(a),
        a["title"].(string),
        storeCookie,
        responseHandlers,
    }

    log.Debugf("WSAction: %v", WSAction)

    return WSAction, valid
}
