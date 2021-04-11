package core

import (
	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/feeder"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"runtime/trace"
	"time"
)

type InjectStruct struct {
	Script string
	Trace bool
	No_log bool
	Disable_dns_cache bool
	Trace_requests bool
	Syntax_check_only bool
	Display_srv_resp bool
	Store_srv_response_dir string
	Listen_addr string
	Output_type string
	Output_dir string
}

func StartStandaloneMode(injectParms InjectStruct) {
	// Read the scenario from file
	data, err := ioutil.ReadFile(injectParms.Script)
	if err != nil {
		log.Fatal(err)
	}

	_startStandaloneMode(
		injectParms.Script,
		data,
		injectParms.No_log,
		injectParms.Disable_dns_cache,
		injectParms.Listen_addr,
		injectParms.Display_srv_resp,
		injectParms.Trace_requests,
		injectParms.Store_srv_response_dir,
		injectParms.Trace,
		injectParms.Syntax_check_only,
		injectParms.Output_dir,
		injectParms.Output_type,
		)
}

func _startStandaloneMode(script_name string, data []byte,
							no_log bool, disable_dns_cache bool, listen_addr string, display_srv_resp bool,
							trace_requests bool, store_srv_response_dir string, must_trace bool, syntax_check_only bool,
							output_dir string, output_type string) {
	log.Info("If you press <Ctrl-C> during the play, you will get partial results !")
	action.DisableLogAction(no_log)
	action.DisableDNSCache(disable_dns_cache)
	action.SetContext(false, listen_addr, display_srv_resp, trace_requests, store_srv_response_dir)

	if must_trace {
		f, err := os.Create("trace.out")
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		err = trace.Start(f)
		if err != nil {
			log.Fatal(err)
		}
		defer trace.Stop()
	}

	// Always creates a Hub for Accept Result in SpawnUsers
	log.Debugf("listen_addr=%s", listen_addr)
	hub = newHub()

	//if !createPlaybook(gp_scriptfile, []byte(data), &gp_playbook, &gp_actions) {
	if !action.CreatePlaybook(script_name, []byte(data), &g_playbook, &g_pre_actions, &g_actions) {
		log.Fatalf("Error while processing the Script File")
	}
	if syntax_check_only {
		log.Info("Syntax check done. Leaving...")
		return
	}

	if g_playbook.DataFeeder.Type == "csv" {
		if !feeder.Csv(g_playbook.DataFeeder, path.Dir(script_name)) {
			return
		}
	} else if g_playbook.DataFeeder.Type != "" {
		log.Fatalf("Unsupported feeder type: %s", g_playbook.DataFeeder.Type)
	}

	reporter.SimulationStart = time.Now()

	outputfile, dir := utils.ComputeOutputFilename(output_dir, output_type)
	if err := reporter.InitReport(output_type); err != nil {
		log.Fatal(err)
	}
	reporter.OpenResultsFile(outputfile)

	go shutdownHandler()
	playPreActions(&g_playbook, &g_pre_actions)
	spawnUsers(&g_playbook, &g_actions, StandaloneMode)

	log.Infof("Done in %v", time.Since(reporter.SimulationStart))
	log.Infof("Building reports, please wait...")
	reporter.CloseResultsFile()
	log.Infof("Count of remaining goroutines=%d", runtime.NumGoroutine())

	err := reporter.CloseReport(outputfile, dir, script_name)
	if err != nil {
		log.Error(err.Error())
	}
	scriptnames := []string{script_name}
	err = reporter.WriteMetadata(reporter.SimulationStart, time.Now(), dir, scriptnames)
	if err != nil {
		log.Error(err.Error())
	}
}

