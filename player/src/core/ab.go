package core

import (
	"fmt"
	"github.com/majeinfo/chaingun/action"
	log "github.com/sirupsen/logrus"
)

type AbStruct struct {
	Trace                  bool
	No_log                 bool
	Disable_dns_cache      bool
	Trace_requests         bool
	Display_srv_resp       bool
	Store_srv_response_dir string
	Listen_addr            string
	Output_dir             string
	Request                string
	Method                 string
	Body                   string
	Iterations             int
	Duration               int
	Users                  int
	Rampup                 int
}

func StartAbMode(abParms AbStruct) {
	/* Initialise the Test Definition from parameters of the command line.
	   Then, check the validity of the test and inject the script.
	*/
	data := fmt.Sprintf(`
iterations: %d
duration: %d
users: %d
rampup: %d
actions:
  - http:
      title: %s
      url: %s
      method: %s
      body: %s
`,
		abParms.Iterations, abParms.Duration, abParms.Users, abParms.Rampup,
		abParms.Request, abParms.Request, abParms.Method, abParms.Body)

	if !action.CreatePlaybook("ab", []byte(data), &g_playbook, &g_pre_actions, &g_actions, &g_post_actions) {
		log.Fatalf("Error while processing the Script File")
	}

	_startStandaloneMode(
		"ab",
		[]byte(data),
		abParms.No_log,
		abParms.Disable_dns_cache,
		abParms.Listen_addr,
		abParms.Display_srv_resp,
		abParms.Trace_requests,
		abParms.Store_srv_response_dir,
		abParms.Trace,
		false,
		abParms.Output_dir,
		"csv",
	)
}
