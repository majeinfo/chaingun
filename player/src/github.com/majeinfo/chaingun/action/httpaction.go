package action

import (
	"regexp"
    log "github.com/sirupsen/logrus"
    "gopkg.in/xmlpath.v2"
    "github.com/majeinfo/chaingun/reporter"
)

type HttpAction struct {
    Method          string              `yaml:"method"`
    Url             string              `yaml:"url"`
    Body            string              `yaml:"body"`
    Template        string              `yaml:"template"`
    Accept          string              `yaml:"accept"`
    ContentType     string              `yaml:"contentType"`
    Title           string              `yaml:"title"`
    ResponseHandler HttpResponseHandler `yaml:"response"`
    StoreCookie     string              `yaml:"storeCookie"`
}

func (h HttpAction) Execute(resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string) {
    DoHttpRequest(h, resultsChannel, sessionMap)
}

type HttpResponseHandler struct {
    Jsonpath string `yaml:"jsonpath"`
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

func NewHttpAction(a map[interface{}]interface{}) HttpAction {
    var valid bool = true
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

    if a["response"] != nil {
        r := a["response"].(map[interface{}]interface{})
        valid_index := []string{"first", "last", "random"}
        if !stringInSlice(r["index"].(string), valid_index) {
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

        // TODO perhaps compile Xmlpath expressions so we can validate early?

        if r["variable"] == nil || r["variable"] == "" {
            log.Error("Error: HttpAction ResponseHandler must define a Variable.")
            valid = false
        }
    }

    if !valid {
        log.Fatalf("Your YAML defintion contains an invalid HttpAction, see errors listed above.")
    }
    
    var responseHandler HttpResponseHandler
    if a["response"] != nil {
        response := a["response"].(map[interface{}]interface{})

        if response["jsonpath"] != nil && response["jsonpath"] != "" {
            responseHandler.Jsonpath = response["jsonpath"].(string)
        }
        if response["xmlpath"] != nil && response["xmlpath"] != "" {
            //responseHandler.Xmlpath = response["xmlpath"].(string)
            var err error
            responseHandler.Xmlpath, err = xmlpath.Compile(response["xmlpath"].(string))
            if err != nil {
				log.Error("XmlPath could not be compiled: %s", response["xmlpath"].(string))
			}
        }
        if response["regex"] != nil && response["regex"] != "" {
			var err error
            responseHandler.Regex, err = regexp.Compile(response["regex"].(string))
            if err != nil {
				log.Error("Regexp could not be compiled: %", response["regex"].(string))
			}
        }

        responseHandler.Variable = response["variable"].(string)
        responseHandler.Index = response["index"].(string)
        responseHandler.Defaultvalue = response["default_value"].(string)
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

    httpAction := HttpAction{
        a["method"].(string),
        a["url"].(string),
        getBody(a),
        getTemplate(a),
        accept,
        contentType,
        a["title"].(string),
        responseHandler,
        storeCookie,
    }

    return httpAction
}
