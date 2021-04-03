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
	action.DisableLogAction(injectParms.No_log)
	action.DisableDNSCache(injectParms.Disable_dns_cache)
	action.SetContext(false, injectParms.Listen_addr, injectParms.Display_srv_resp, injectParms.Trace_requests, injectParms.Store_srv_response_dir)

	if injectParms.Trace {
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
	log.Debugf("listen_addr=%s", injectParms.Listen_addr)
	hub = newHub()

	// Read the scenario from file
	data, err := ioutil.ReadFile(injectParms.Script)
	if err != nil {
		log.Fatal(err)
	}

	//if !createPlaybook(gp_scriptfile, []byte(data), &gp_playbook, &gp_actions) {
	if !action.CreatePlaybook(injectParms.Script, []byte(data), &g_playbook, &g_pre_actions, &g_actions) {
		log.Fatalf("Error while processing the Script File")
	}
	if injectParms.Syntax_check_only {
		log.Info("Syntax check done. Leaving...")
		return
	}

	if g_playbook.DataFeeder.Type == "csv" {
		if !feeder.Csv(g_playbook.DataFeeder, path.Dir(injectParms.Script)) {
			return
		}
	} else if g_playbook.DataFeeder.Type != "" {
		log.Fatalf("Unsupported feeder type: %s", g_playbook.DataFeeder.Type)
	}

	reporter.SimulationStart = time.Now()

	outputfile, dir := utils.ComputeOutputFilename(injectParms.Output_dir, injectParms.Output_type)
	if err := reporter.InitReport(injectParms.Output_type); err != nil {
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

	err = reporter.CloseReport(outputfile, dir, injectParms.Script)
	if err != nil {
		log.Error(err.Error())
	}
	scriptnames := []string{injectParms.Script}
	err = reporter.WriteMetadata(reporter.SimulationStart, time.Now(), dir, scriptnames)
	if err != nil {
		log.Error(err.Error())
	}
}

