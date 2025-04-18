package action

import (
	"errors"
	"net/url"
	"strings"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// HTTPAction describes a HTTP Action
type HTTPAction struct {
	Method           string            `yaml:"method"`
	UseHTTP2         bool              `yaml:"use_http2"`
	URL              string            `yaml:"url"`
	Body             string            `yaml:"body"`
	Template         string            `yaml:"template"`
	FormDatas        []FormData        `yaml:"formdatas"`
	Headers          map[string]string `yaml:"headers"`
	Title            string            `yaml:"title"`
	UploadFile       []byte            `yaml:"upload_file"`
	StoreCookie      string            `yaml:"store_cookie"`
	ResponseHandlers []ResponseHandler `yaml:"responses"`
}

// FormData describes the data that will be sent with the HTTP Request
type FormData struct {
	Name    string `yaml:"name"`
	Value   string `yaml:"value"`
	Type    string `yaml:"type"`
	Content []byte
}

var (
	errFormDataBlockAnalysis = errors.New("Errors occurred during FormData block analysis")
)

// Execute a HTTP Action
func (h HTTPAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	vulog.Data["action"] = h.Title
	return DoHTTPRequest(h, resultsChannel, sessionMap, vucontext, vulog, playbook)
}

// NewHTTPAction creates a new HTTP Action
func NewHTTPAction(a map[interface{}]interface{}, dflt config.Default, playbook *config.TestDef) (HTTPAction, bool) {
	log.Debugf("NewhttpAction=%v", a)
	valid := true

	if a["url"] == "" || a["url"] == nil {
		log.Error("HttpAction must define a URL.")
		a["url"] = ""
		valid = false
	} else {
		// Try to substitute already known variables: needed if variables are used
		// protocol://in the user:auth@server:port/ part of the URL
		// (cannot use SubstParams() here)
		// TODO: why here and not in DoHTTPRequest ? (same question for Mongo, SQL, etc...)
		textData := a["url"].(string)
		if strings.ContainsAny(textData, "${") {
			res := re.FindAllStringSubmatch(textData, -1)
			for _, v := range res {
				log.Debugf("playbook.Variables[%s]=%s", v[1], playbook.Variables[v[1]])
				if _, err := playbook.Variables[v[1]]; !err {
					log.Debugf("Variable ${%s} not set", v[1])
				} else {
					textData = strings.Replace(textData, "${"+v[1]+"}", url.QueryEscape(playbook.Variables[v[1]].Values[0]), 1) // TODO array
				}
			}
			a["url"] = textData
		}
		valid = setDefaultURL(a, dflt)
		log.Debugf("setDefaultURL returned %v", a)
	}

	if a["method"] == nil || a["method"] == "" {
		if dflt.Method == "" {
			log.Error("Action has no Method and no default Method specified")
			a["method"] = ""
			valid = false
		} else {
			a["method"] = dflt.Method
		}
	} else if _, err := config.IsValidHTTPMethod(a["method"].(string)); err != nil {
		log.Errorf("%v", err)
		valid = false
	}
	if a["title"] == nil || a["title"] == "" {
		log.Error("HttpAction must define a title.")
		a["title"] = ""
		valid = false
	}
	if a["use_http2"] == nil {
		a["use_http2"] = false
	} else {
		if _, ok := a["use_http2"].(bool); !ok {
			log.Error("use_http2 value must be a boolean (true or false)")
			a["use_http2"] = false
			valid = false
		}
	}

	// Check formdatas
	nu := 0
	if a["body"] != nil {
		nu++
	}
	if a["template"] != nil {
		addEmbeddedFilename(a["template"].(string))
		nu++
	}
	if a["upload_file"] != nil {
		addEmbeddedFilename(a["upload_file"].(string))
		nu++
	}
	if a["formdata"] != nil {
		nu++
	}
	if nu > 1 {
		log.Error("A HttpAction can contain a single 'body' or a 'template' or a 'formdata' or an 'upload_file'.")
		valid = false
	}

	var storeCookie string
	if a["store_cookie"] != nil && a["store_cookie"].(string) != "" {
		storeCookie = a["store_cookie"].(string)
	}

	headers := make(map[string]string, 20)
	if a["headers"] != nil {
		// Check the type : otherwise crashes if headers content is a list instead of a map...
		switch v := a["headers"].(type) {
		case map[interface{}]interface{}:
			//for hdr, value := range a["headers"].(map[interface{}]interface{}) {
			for hdr, value := range v {
				log.Debugf("Header Key=%s / Value=%s", hdr.(string), value.(string))
				headers[strings.ToLower(hdr.(string))] = value.(string)
			}
		default:
			log.Fatalf("headers format is invalid: it should be a map (you probably set it as a list ?)")
		}
	}

	// Set the Accept header if not set in Playbook
	if _, ok := headers["accept"]; !ok {
		headers["accept"] = "text/html,application/json,application/xhtml+xml,application/xml,text/plain"
	}
	// Set the User-Agent header if not set in Playbook
	if _, ok := headers["user-agent"]; !ok {
		if is_daemon_mode {
			headers["user-agent"] = "chaingun-" + injector_id
		} else {
			headers["user-agent"] = "chaingun"
		}
	}

	formdatas, validData := NewFormDatas(a)
	responseHandlers, validResp := NewResponseHandlers(a)
	template, validTempl := getTemplate(a)
	body, validBody := getBody(a)
	upload, validUpload := getFileToPUT(a)

	if !valid || !validResp || !validData || !validTempl || !validBody || !validUpload {
		log.Errorf("Your YAML Playbook contains an invalid HTTPAction, see errors listed above.")
		valid = false
	}

	httpAction := HTTPAction{
		Method:           a["method"].(string),
		UseHTTP2:         a["use_http2"].(bool),
		URL:              a["url"].(string),
		Body:             body,
		Template:         template,
		FormDatas:        formdatas,
		Headers:          headers,
		Title:            a["title"].(string),
		UploadFile:       upload,
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
			log.Error("formdata format is invalid: should be a list of maps")
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
			addEmbeddedFilename(formData.Value)
			formData.Content, valid = getFileToUpload(formData.Value)
		}
	}

	if !valid {
		return formData, errFormDataBlockAnalysis
	}

	return formData, nil
}
