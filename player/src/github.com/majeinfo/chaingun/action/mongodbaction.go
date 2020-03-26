package action

import (
	_ "errors"
	"net/url"
	"strings"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// MongoDBAction describes a MongoDB Action
type MongoDBAction struct {
	Server           string            `yaml:"server"`
	Title            string            `yaml:"title"`
	Database         string            `yaml:"database"`
	Collection       string            `yaml:"collection"`
	Command          string            `yaml:"command"`
	Document         string            `yaml:"document"`
	Filter           string            `yaml:"filter"`
	ResponseHandlers []ResponseHandler `yaml:"responses"`
}

// Execute a MongoDB Action
func (h MongoDBAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	vulog.Data["action"] = h.Title
	return DoMongoDBRequest(h, resultsChannel, sessionMap, vulog, playbook)
}

// NewMongoDBAction creates a new MongoDB Action
func NewMongoDBAction(a map[interface{}]interface{}, dflt config.Default, playbook *config.TestDef) (MongoDBAction, bool) {
	log.Debugf("NewMongoDBAction=%v", a)
	valid := true

	if a["server"] == "" || a["server"] == nil {
		if dflt.Server == "" {
			log.Error("MongoDBAction has no Server and no default Server specified")
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
					textData = strings.Replace(textData, "${"+v[1]+"}", url.QueryEscape(playbook.Variables[v[1]]), 1)
				}
			}
			a["server"] = textData
		}
		//valid = setDefaultURL(a, dflt)
		//log.Debugf("setDefaultURL returned %v", a)
	}

	if a["command"] == nil || a["command"] == "" {
		log.Error("MongoDBAction has no Command and no default Command specified")
		a["command"] = ""
		valid = false
	} else if !config.IsValidMongoDBCommand(a["command"].(string)) {
		log.Error("MongoDBAction must specify a valid command: insertone, findone, deletemany, drop")
		valid = false
	}

	if a["command"] == "insertone" {
		if a["document"] == nil || a["document"] == "" {
			log.Error("MongoDBAction insertone command must define a document.")
			a["document"] = ""
			valid = false
		}
	} else {
		a["document"] = ""
	}

	if a["command"] == "findone" {
		if a["filter"] == nil || a["filter"] == "" {
			log.Error("MongoDBAction findone command must define a filter.")
			a["filter"] = ""
			valid = false
		}
	} else {
		a["filter"] = ""
	}

	if a["title"] == nil || a["title"] == "" {
		log.Error("MongoDBAction must define a title.")
		a["title"] = ""
		valid = false
	}

	if a["database"] == nil || a["database"] == "" {
		if dflt.Database == "" {
			log.Error("MongoDBAction has no Database and no default Database specified")
			a["database"] = ""
			valid = false
		} else {
			a["database"] = dflt.Database
		}
	}

	if a["collection"] == nil || a["collection"] == "" {
		if dflt.Collection == "" {
			log.Error("MongoDBAction has no Collection and no default Collection specified")
			a["collection"] = ""
			valid = false
		} else {
			a["collection"] = dflt.Collection
		}
	}

	responseHandlers, validResp := NewResponseHandlers(a)

	if !valid || !validResp {
		log.Errorf("Your YAML Playbook contains an invalid MongoDBAction, see errors listed above.")
		valid = false
	}

	mongodbAction := MongoDBAction{
		Server:           a["server"].(string),
		Title:            a["title"].(string),
		Database:         a["database"].(string),
		Collection:       a["collection"].(string),
		Command:          a["command"].(string),
		Document:         a["document"].(string),
		Filter:           a["filter"].(string),
		ResponseHandlers: responseHandlers,
	}

	log.Debugf("MongoDBAction: %v", mongodbAction)

	return mongodbAction, valid
}
