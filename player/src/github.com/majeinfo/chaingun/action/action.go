package action

import (
	"fmt"
	"os"
	"strings"

	"github.com/Knetic/govaluate"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// Action is an interface which is able to execute a Request
type Action interface {
	// Returns false if an error occurred during the execution
	Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vucontext *config.VUContext, vulog *log.Entry, playbook *config.TestDef) bool
}

// FullAction embeds the global parameters for all actions as well as an Action
type FullAction struct {
	When         string `yaml:"when"`
	CompiledWhen *govaluate.EvaluableExpression
	Action       Action
}

var (
	gpScriptDir           string
	is_daemon_mode        bool
	injector_id           string // unique ID for an injector
	must_display_srv_resp bool
	must_trace_request    bool
	store_srv_resp_dir    string
	disable_dns_cache     bool
	embedded_files        []string // list of filenames embedded in the current playbook
)

func SetContext(daemon_mode bool, injectorID string, displaySrvResp bool, mustTraceReq bool, storeSrvRespDir string) {
	is_daemon_mode = daemon_mode
	injector_id = injectorID
	must_display_srv_resp = displaySrvResp
	must_trace_request = mustTraceReq
	store_srv_resp_dir = storeSrvRespDir

	if store_srv_resp_dir != "" {
		// Creates the outputdir if needed
		stat, err := os.Stat(store_srv_resp_dir)
		if os.IsNotExist(err) {
			log.Debugf("Must create the Server Response Output Directory %s", store_srv_resp_dir)
			if err := os.MkdirAll(store_srv_resp_dir, 0755); err != nil {
				log.Fatalf("Cannot create Server Response Output Directory %s: %s", store_srv_resp_dir, err.Error())
			}
		} else if !stat.Mode().IsDir() {
			log.Fatalf("Server Response Output Directory %s already exists as a file !", store_srv_resp_dir)
		}
	}
}

func DisableDNSCache(discache bool) {
	disable_dns_cache = discache
}

// Store the whole response in a directory.
// The filename is built from the action's "title", the VU number and the number of iteration
func store_srv_resp(title string, UID string, iter_nu int, response []byte) {
	if store_srv_resp_dir == "" {
		return
	}

	log.Debugf("Title=%s, VU=%s, Iter_nu=%d", title, UID, iter_nu)

	// Compute filename, create it and write data !
	filename := fmt.Sprintf("%s/%s-VU-%s-#%d", store_srv_resp_dir, strings.Replace(title, " ", "_", -1), UID, iter_nu)
	file, err := os.Create(filename)
	if err != nil {
		log.Errorf("Couldn't open response file %s: %s", filename, err)
	} else {
		defer file.Close()
		file.Write(response)
	}
}
