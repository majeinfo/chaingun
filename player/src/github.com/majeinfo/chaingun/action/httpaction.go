package action

import (
    "errors"
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
    FormDatas       []FormData          `yaml:"formdatas"`
    Headers         map[string]string   `yaml:"headers"`
    Title           string              `yaml:"title"`
    StoreCookie     string              `yaml:"storeCookie"`
    ResponseHandlers []ResponseHandler  `yaml:"responses"`
}

// These data will be sent with
type FormData struct {
    Name    string      `yaml:"name"`
    Value   string      `yaml:"name"`
    Type    string      `yaml:"type"`
    Content []byte
}

func (h HttpAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
    return DoHttpRequest(h, resultsChannel, sessionMap, playbook)
}

func NewHttpAction(a map[interface{}]interface{}, dflt config.Default) (HttpAction, bool) {
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

    // Check formdatas
    nu := 0
    if a["body"] != nil { nu ++ }
    if a["template"] != nil { nu++ }
    if a["formdata"] != nil { nu++ }
    if nu > 1 {
        log.Error("A HttpAction can contain a single 'body' or a 'template' or a 'formdata'.")
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

    formdatas, validData := NewFormDatas(a)
    responseHandlers, validResp  := NewResponseHandlers(a)
    template, validTempl := getTemplate(a)

    if !valid || !validResp || !validData || !validTempl {
        log.Errorf("Your YAML Playbook contains an invalid HTTPAction, see errors listed above.")
        valid = false
    }

    httpAction := HttpAction{
        a["method"].(string),
        a["url"].(string),
        getBody(a),
        template,
        formdatas,
        headers,
        a["title"].(string),
        storeCookie,
        responseHandlers,
    }

    log.Debugf("HTTPAction: %v", httpAction)

    return httpAction, valid
}

// Build all the FormDatas from the Action described in YAML Playbook
func NewFormDatas(a map[interface{}]interface{}) ([]FormData, bool) {
	valid := true
	var formDatas []FormData
    if a["formdata"] == nil {
        formDatas = nil
    } else {
        switch v := a["formdata"].(type) {
        case []interface {}:
            formDatas = make([]FormData, len(v))
            for idx, r1 := range v {
                r2 := r1.(map[interface{}]interface{})
                log.Debugf("formdata=%v", r2)
                newFormData, err := NewFormData(r2)
                if err != nil {
                    valid = false
                    break
                }
                formDatas[idx] = newFormData
            }
        default:
            log.Error("formdata format is invalid")
            valid = false
        }
	}
	
	return formDatas, valid
}

func NewFormData(a map[interface{}]interface{}) (FormData, error) {
	valid := true
	var formData FormData
        
    if a["name"] == nil {
        log.Error("FormData must have a 'name' attribute.")
        valid = false
    } else {
        formData.Name = a["name"].(string)
    }
    if a["value"] == nil {
        log.Error("FormData must have a 'value' attribute.")
        valid = false
    } else {
        formData.Value = a["value"].(string)
    }

    if a["type"] != nil && valid {
        formData.Type = a["type"].(string)
        if formData.Type != "file" {
            log.Error("'type' attribute of FormData must be 'file'.")
            valid = false
        } else {
            formData.Content, valid = getFileToUpload(formData.Value)
        }
    }

	if !valid {
		return formData, errors.New("Errors occurred during FormData block analysis.")
	}

    return formData, nil
}