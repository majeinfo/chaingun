package action

import (
	_ "errors"
	"net/url"
	"strings"

	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// SQLAction describes a SQL Action
type SQLAction struct {
	DBDriver         string            `yaml:"db_driver"`
	Server           string            `yaml:"server"`
	Title            string            `yaml:"title"`
	Database         string            `yaml:"database"`
	Statement        string            `yaml:"statement"`
	ResponseHandlers []ResponseHandler `yaml:"responses"`
}

// Execute a SQL Action
func (h SQLAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	vulog.Data["action"] = h.Title
	return DoSQLRequest(h, resultsChannel, sessionMap, vucontext, vulog, playbook)
}

// NewSQLAction creates a new MongoDB Action
func NewSQLAction(a map[interface{}]interface{}, dflt config.Default, playbook *config.TestDef) (SQLAction, bool) {
	log.Debugf("NewSQLAction=%v", a)
	valid := true

	if a["db_driver"] == "" || a["db_driver"] == nil {
		if dflt.DBDriver == "" {
			log.Error("SQLAction has no Driver and no default Driver specified")
			a["db_driver"] = ""
			valid = false
		} else {
			a["db_driver"] = dflt.DBDriver
		}
	} else {
		if !config.IsValidDBDriver(a["db_driver"].(string)) {
			log.Error("DB Driver must specify a valid driver (mysql)")
			valid = false
		}
	}

	if a["server"] == "" || a["server"] == nil {
		if dflt.Server == "" {
			log.Error("SQLAction has no Server and no default Server specified")
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
	}

	if a["title"] == nil || a["title"] == "" {
		log.Error("SQLAction must define a title")
		a["title"] = ""
		valid = false
	}

	if a["database"] == nil || a["database"] == "" {
		if dflt.Database == "" {
			log.Error("SQLAction has no Database and no default Database specified")
			a["database"] = ""
			valid = false
		} else {
			a["database"] = dflt.Database
		}
	}

	if a["statement"] == nil || a["statement"] == "" {
		log.Error("SQLAction has no Statement specified")
		a["statement"] = ""
		valid = false
	}

	responseHandlers, validResp := NewResponseHandlers(a)

	if !valid || !validResp {
		log.Errorf("Your YAML Playbook contains an invalid SQLAction, see errors listed above")
		valid = false
	}

	sqlAction := SQLAction{
		DBDriver:         a["db_driver"].(string),
		Server:           a["server"].(string),
		Title:            a["title"].(string),
		Database:         a["database"].(string),
		Statement:        a["statement"].(string),
		ResponseHandlers: responseHandlers,
	}

	log.Debugf("SQLAction: %v", sqlAction)

	return sqlAction, valid
}
