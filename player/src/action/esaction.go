package action

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// ESAction describes an ElasticSearch Action
// TODO: should support Bearer Token ?
type ESAction struct {
	Server           string            `yaml:"server"` // may also specify the username and the password
	Title            string            `yaml:"title"`
	Command          string            `yaml:"command"`
	Index            string            `yaml:"index"`
	Document         string            `yaml:"document"`
	Query            string            `yaml:"query"`
	Refresh		 bool		   `yaml:"refresh"`	// for "insert" action (default is false)
	ResponseHandlers []ResponseHandler `yaml:"responses"`
}

const (
	REPORTER_ES    string = "ELASTICSEARCH"
	ES_CREATEINDEX        = "createindex"
	ES_DELETEINDEX        = "deleteindex"
	ES_INSERT             = "insert"
	ES_SEARCH             = "search"
)

// Execute an ElasticSearch Action
func (h ESAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	vulog.Data["action"] = h.Title
	return DoESRequest(h, resultsChannel, sessionMap, vucontext, vulog, playbook)
}

// NewESAction creates a new ESAction
func NewESAction(a map[interface{}]interface{}, dflt config.Default, playbook *config.TestDef) (ESAction, bool) {
	log.Debugf("NewESAction=%v", a)
	valid := true

	if a["server"] == "" || a["server"] == nil {
		if dflt.Server == "" {
			log.Error("ESAction has no Server and no default Server specified")
			a["server"] = ""
			valid = false
		} else {
			a["server"] = dflt.Server
		}
	} else {
		// Try to substitute already known variables: needed if variables are used
		// protocol://in the user:auth@server:port/ part of the URL
		// (cannot use SubstParams() here)
		textData := a["server"].(string)
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
			a["server"] = textData
		}
		//valid = setDefaultURL(a, dflt)
		//log.Debugf("setDefaultURL returned %v", a)
	}

	if a["command"] == nil || a["command"] == "" {
		log.Error("ESAction has no Command and no default Command specified")
		a["command"] = ""
		valid = false
	} else if _, err := isValidESCommand(a["command"].(string)); err != nil {
		log.Errorf("%v", err)
		valid = false
	}

	if a["command"] == "insert" {
		if a["document"] == nil || a["document"] == "" {
			log.Error("ESAction insert command must define a document")
			a["document"] = ""
			valid = false
		}
	} else {
		a["document"] = ""
	}

	if a["command"] == "search" {
		if a["query"] == nil || a["query"] == "" {
			log.Error("ESAction search command must define a query")
			a["query"] = ""
			valid = false
		}
	} else {
		a["query"] = ""
	}

	if a["title"] == nil || a["title"] == "" {
		log.Error("ESAction must define a title")
		a["title"] = ""
		valid = false
	}

	if a["index"] == nil || a["index"] == "" {
		if dflt.Index == "" {
			log.Error("ESAction has no Index specified")
			a["index"] = ""
			valid = false
		} else {
			a["index"] = dflt.Index
		}
	}

        if a["refresh"] == nil {
                a["refresh"] = false
        } else {
                if _, ok := a["refresh"].(bool); !ok {
                        log.Error("refresh value must be a boolean (true or false)")
                        a["refresh"] = false
                        valid = false
                }
        }

	responseHandlers, validResp := NewResponseHandlers(a)

	if !valid || !validResp {
		log.Errorf("Your YAML Playbook contains an invalid ESAction, see errors listed above")
		valid = false
	}

	esAction := ESAction{
		Server:           a["server"].(string),
		Title:            a["title"].(string),
		Command:          a["command"].(string),
		Index:            a["index"].(string),
		Document:         a["document"].(string),
		Query:            a["query"].(string),
		Refresh:           a["refresh"].(bool),
		ResponseHandlers: responseHandlers,
	}

	log.Debugf("ESAction: %v", esAction)

	return esAction, valid
}

func isValidESCommand(command string) (bool, error) {
	valid_commands := []string{ES_CREATEINDEX, ES_INSERT, ES_SEARCH, ES_DELETEINDEX}

	if !config.StringInSlice(command, valid_commands) {
		return false, fmt.Errorf("ESAction must specify a valid command: createindex, deleteindex, insert, search: got %s", command)
	}

	return true, nil
}
