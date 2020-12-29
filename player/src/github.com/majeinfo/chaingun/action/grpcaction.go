package action

import (
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// HTTPAction describes a HTTP Action
type GRPCAction struct {
	Title            string            `yaml:"title"`
	Call           string            `yaml:"call"`
	Data string                `yaml:"data"`
}


// Execute a GRPC Action
func (h GRPCAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool {
	vulog.Data["action"] = h.Title
	return DoGRPCRequest(h, resultsChannel, sessionMap, vucontext, vulog, playbook)
}

// NewHTTPAction creates a new HTTP Action
func NewGRPCAction(a map[interface{}]interface{}, dflt config.Default, playbook *config.TestDef) (GRPCAction, bool) {
	log.Debugf("NewGRPCAction=%v", a)
	valid := true

	if a["call"] == "" || a["call"] == nil {
		log.Error("GrpcAction must define a function call name.")
		a["call"] = ""
		valid = false
	}

	if a["title"] == nil || a["title"] == "" {
		log.Error("GrpcAction must define a title.")
		a["title"] = ""
		valid = false
	}

	if !valid  {
		log.Errorf("Your YAML Playbook contains an invalid GRPCAction, see errors listed above.")
		valid = false
	}

	grpcAction := GRPCAction{
		Title:            a["title"].(string),
		Call: a["call"].(string),
		Data: a["data"].(string),
	}

	log.Debugf("GRPCAction: %v", grpcAction)

	return grpcAction, valid
}
