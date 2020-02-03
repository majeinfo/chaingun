package action

import (
	"github.com/Knetic/govaluate"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// Action is an interface which is able to execute a Request
type Action interface {
	// Returns false if an error occurred during the execution
	Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool
}

// FullAction embeds the global parameters for all actions as well as an Action
type FullAction struct {
	When         string `yaml:"when"`
	CompiledWhen *govaluate.EvaluableExpression
	Action       Action
}

var (
	is_daemon_mode        bool
	injector_id           string // unique ID for an injector
	must_display_srv_resp bool
	must_trace_request    bool
	disable_dns_cache     bool
	embedded_files        []string // list of filenames embedded in the current playbook
)

func SetContext(daemon_mode bool, injectorID string, displaySrvResp bool, mustTraceReq bool) {
	is_daemon_mode = daemon_mode
	injector_id = injectorID
	must_display_srv_resp = displaySrvResp
	must_trace_request = mustTraceReq
}

func DisableDNSCache(discache bool) {
	disable_dns_cache = discache
}
