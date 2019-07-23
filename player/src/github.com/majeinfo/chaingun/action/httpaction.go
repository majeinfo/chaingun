package action

import (
	"errors"
	"strings"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// HTTPAction describes a HTTP Action
type HTTPAction struct {
	Method           string            `yaml:"method"`
	URL              string            `yaml:"url"`
	Body             string            `yaml:"body"`
	Template         string            `yaml:"template"`
	FormDatas        []FormData        `yaml:"formdatas"`
	Headers          map[string]string `yaml:"headers"`
	Title            string            `yaml:"title"`
	StoreCookie      string            `yaml:"storeCookie"`
	ResponseHandlers []ResponseHandler `yaml:"responses"`
}

// FormData describes the data that will be sent with the HTTP Request
type FormData struct {
	Name    string `yaml:"name"`
	Value   string `yaml:"value"`
	Type    string `yaml:"type"`
	Content []byte
}

// Execute a HTTP Action
func (h HTTPAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	vulog.Data["action"] = h.Title
	return DoHTTPRequest(h, resultsChannel, sessionMap, vulog, playbook)
}

// NewHTTPAction creates a new HTTP Action
func NewHTTPAction(a map[interface{}]interface{}, dflt config.Default) (HTTPAction, bool) {
	log.Debugf("NewhttpAction=%v", a)
	valid := true

	if a["url"] == "" || a["url"] == nil {
		log.Error("HttpAction must define a URL.")
		a["url"] = ""
		valid = false
	} else {
		valid = setDefaultURL(a, dflt)
	}

	if a["method"] == nil || a["method"] == "" {
		if dflt.Method == "" {
			log.Error("Action has no Method and no default Method specified")
			a["method"] = ""
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
		a["title"] = ""
		valid = false
	}

	// Check formdatas
	nu := 0
	if a["body"] != nil {
		nu++
	}
	if a["template"] != nil {
		nu++
	}
	if a["formdata"] != nil {
		nu++
	}
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
		for hdr, value := range a["headers"].(map[interface{}]interface{}) {
			log.Debugf("Header Key=%s / Value=%s", hdr.(string), value.(string))
			headers[strings.ToLower(hdr.(string))] = value.(string)
		}
	}
	if _, ok := headers["accept"]; !ok {
		headers["accept"] = "text/html,application/json,application/xhtml+xml,application/xml,text/plain"
	}
	headers["user-agent"] = "chaingun-by-JD"

	formdatas, validData := NewFormDatas(a)
	responseHandlers, validResp := NewResponseHandlers(a)
	template, validTempl := getTemplate(a)

	if !valid || !validResp || !validData || !validTempl {
		log.Errorf("Your YAML Playbook contains an invalid HTTPAction, see errors listed above.")
		valid = false
	}

	httpAction := HTTPAction{
		Method:           a["method"].(string),
		URL:              a["url"].(string),
		Body:             getBody(a),
		Template:         template,
		FormDatas:        formdatas,
		Headers:          headers,
		Title:            a["title"].(string),
		StoreCookie:      storeCookie,
		ResponseHandlers: responseHandlers,
	}

	log.Debugf("HTTPAction: %v", httpAction)

	return httpAction, valid
}

// NewFormDatas builds all the FormDatas from the Action described in YAML Playbook
func NewFormDatas(a map[interface{}]interface{}) ([]FormData, bool) {
	valid := true
	var formDatas []FormData
	if a["formdata"] == nil {
		formDatas = nil
	} else {
		switch v := a["formdata"].(type) {
		case []interface{}:
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

// NewFormData build a new structure to handle form data to be sent
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
		return formData, errors.New("Errors occurred during FormData block analysis")
	}

	return formData, nil
}
