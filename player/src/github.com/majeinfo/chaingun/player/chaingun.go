package main

import (
	"io/ioutil"
	_ "math/rand"
	_ "net"
	"os"
	"path"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/feeder"
	"github.com/majeinfo/chaingun/manager"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// Program starts here
func main() {
	command_line()

	//runtime.SetMutexProfileFraction(1)
	//runtime.SetBlockProfileRate(1)

	if *gp_cpu_profile != "" {
		f, err := os.Create(*gp_cpu_profile)
		if err != nil {
			log.Fatal("Could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("Could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	gp_mode.f()

	if *gp_mem_profile != "" {
		f, err := os.Create(*gp_mem_profile)
		if err != nil {
			log.Fatal("Could not create memory profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		runtime.GC()    // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("Could not write memory profile: ", err)
		}
		/*
			if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
				log.Fatal("Could not write memory profile: ", err)
			}
		*/
	}
}

func playStandaloneMode() {
	if *gp_trace {
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
	log.Debugf("*gp_listen_addr=%s", *gp_listen_addr)
	hub = newHub()

	// Read the scenario from file
	data, err := ioutil.ReadFile(*gp_scriptfile)
	if err != nil {
		log.Fatal(err)
	}

	//if !createPlaybook(gp_scriptfile, []byte(data), &gp_playbook, &gp_actions) {
	if !action.CreatePlaybook(gp_scriptfile, []byte(data), &gp_playbook, &gp_pre_actions, &gp_actions) {
		log.Fatalf("Error while processing the Script File")
	}
	if *gp_syntax_check_only {
		log.Info("Syntax check done. Leaving...")
		return
	}

	if gp_playbook.DataFeeder.Type == "csv" {
		if !feeder.Csv(gp_playbook.DataFeeder, path.Dir(*gp_scriptfile)) {
			return
		}
	} else if gp_playbook.DataFeeder.Type != "" {
		log.Fatalf("Unsupported feeder type: %s", gp_playbook.DataFeeder.Type)
	}

	reporter.SimulationStart = time.Now()

	outputfile, dir := computeOutputFilename()
	if err := reporter.InitReport(*gp_outputtype); err != nil {
		log.Fatal(err)
	}
	reporter.OpenResultsFile(outputfile)

	playPreActions(&gp_playbook, &gp_pre_actions)
	spawnUsers(&gp_playbook, &gp_actions)

	log.Infof("Done in %v", time.Since(reporter.SimulationStart))
	log.Infof("Building reports, please wait...")
	reporter.CloseResultsFile()
	log.Infof("Count of remaining goroutines=%d", runtime.NumGoroutine())

	err = reporter.CloseReport(outputfile, dir, *gp_scriptfile)
	if err != nil {
		log.Error(err.Error())
	}
	scriptnames := []string{*gp_scriptfile}
	err = reporter.WriteMetadata(reporter.SimulationStart, time.Now(), dir, scriptnames)
	if err != nil {
		log.Error(err.Error())
	}
}

func playGraphOnlyMode() {
	// Just the graph production
	outputfile, dir := computeOutputFilename()
	if err := reporter.InitReport(*gp_outputtype); err != nil {
		log.Fatal(err)
	}

	if err := reporter.CloseReport(outputfile, dir, *gp_scriptfile); err != nil {
		log.Error(err.Error())
	}
}

func playDaemonMode() {
	// Always creates a Hub for Accept Result in SpawnUsers
	log.Debugf("*gp_listen_addr=%s", *gp_listen_addr)
	hub = newHub()

	if *gp_listen_addr != "" {
		log.Debugf("Create server listening on: %s", *gp_listen_addr)
		startWsServer(*gp_listen_addr)
	} else {
		/*
			conn, err := net.Dial("tcp", *gp_connect_to)
			if err != nil {
				log.Fatalf("Could not connect to %s: %s", *gp_connect_to, err)
			}
		*/
		log.Fatal("connect-to mode is not yet implemented")
	}
}

func playManagerMode() {
	log.Infof("Start manager mode on this address: %s", *gp_manager_addr)
	manager.Start(gp_manager_addr, gp_repositorydir, gp_injectors)
}

func playDesignerMode() {
	log.Infof("Start designer mode on this address: %s", *gp_listen_addr)
	startDesignerMode(gp_listen_addr)
}

func playBatchMode() {
	log.Debug("Batch mode started")
	manager.StartBatch(gp_manager_addr, gp_repositorydir, gp_injectors, gp_scriptfile)
}

// Compute the name of the output file (/path/to/data.csv)
func computeOutputFilename() (string, string) {
	var outputfile string
	var dir string

	if *gp_outputdir == "" {
		d, _ := os.Getwd()
		dir = d + "/results"
	} else {
		dir = *gp_outputdir
	}
	if dir[len(dir)-1] != '/' {
		dir += "/"
	}
	outputfile = dir + "data." + *gp_outputtype

	return outputfile, dir
}

// EOF
