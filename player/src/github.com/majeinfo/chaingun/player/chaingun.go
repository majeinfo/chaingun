package main

import (
	"flag"
	"io/ioutil"
	_ "math/rand"
	_ "net"
	"os"
	"path"
	"strconv"
	"sync"
	"time"

	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/feeder"
	"github.com/majeinfo/chaingun/manager"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	standaloneMode = 0 + iota
	daemonMode
	managerMode
)

var (
	modeTypeMap = map[string]int{
		"standalone": standaloneMode,
		"daemon":     daemonMode,
		"manager":    managerMode,
	}
)

var (
	VU_start          time.Time
	VU_count          int
	gp_mode           int
	gp_valid_playbook bool = false
	gp_listen_addr    *string
	gp_manager_addr   *string
	gp_connect_to     *string
	gp_scriptfile     *string
	gp_outputdir      *string
	gp_outputtype     *string
	gp_python_cmd     *string
	gp_viewerfile     *string
	gp_no_log         *bool

	gp_playbook config.TestDef
	gp_actions  []action.Action
)

// Analyze the command line
func command_line() {
	mode := flag.String("mode", "standalone", "standalone(default)|daemon|manager")
	gp_listen_addr = flag.String("listen-addr", "127.0.0.1:12345", "Address and port to listen to in daemon")
	gp_manager_addr = flag.String("manager-listen-addr", "127.0.0.1:8000", "Address and port to listen to - for the Manager Web Interface")
	gp_connect_to = flag.String("connect-to", "", "Address and port to connect to - in daemon mode (not supported yet)")
	verbose := flag.Bool("verbose", false, "Set verbose mode")
	gp_scriptfile = flag.String("script", "", "Set the Script")
	gp_outputdir = flag.String("output-dir", "", "Set the output directory")
	gp_outputtype = flag.String("output-type", "csv", "Set the output type in file (csv/default, json)")
	gp_python_cmd = flag.String("python-cmd", "", "Select the Python Interpreter to create the graphs")
	gp_viewerfile = flag.String("viewer", "", "Give the location of viewer.py script")
	gp_no_log = flag.Bool("no-log", false, "Disable the 'log' actions from the Script")

	flag.Parse()

	log_level := log.InfoLevel
	if *verbose {
		log_level = log.DebugLevel
	}
	log.SetLevel(log_level)
	action.DisableAction(*gp_no_log)

	// Check the mode
	var ok bool
	gp_mode, ok = modeTypeMap[*mode]
	if !ok {
		log.Fatalf("Unknown mode value: %s (allowed values are: standalone, daemon or manager)", *mode)
	}
	log.Debugf("Player mode is %s", *mode)

	// Do some command line consistency tests
	if gp_mode == standaloneMode {
		if *gp_scriptfile == "" {
			log.Fatal("When started in standalone mode, needs a 'script' file")
		}
		if *gp_python_cmd == "" {
			*gp_python_cmd = os.Getenv("Python")
			if *gp_python_cmd == "" {
				log.Fatalf("You must specify a Python interpreter path with --python-cmd option or via the PYTHON environment variable")
			}
		}
		if _, err := os.Stat(*gp_python_cmd); os.IsNotExist(err) {
			log.Fatalf("Python interpreter %s does not exist.", *gp_python_cmd)
		}
		if *gp_viewerfile == "" {
			log.Fatal("When started in standalone or manager mode, needs the location of the viewer.py script")
		}
		if _, err := os.Stat(*gp_viewerfile); os.IsNotExist(err) {
			log.Fatalf("The specified Viewer %s does not exist.", *gp_viewerfile)
		}
	}
	if gp_mode == daemonMode {
		// Either listen-addr or connect-to must be specified
		// WebSocket server must not be started
		if *gp_listen_addr != "" && *gp_connect_to != "" {
			log.Fatal("Either --connect-to or --listen-addr options can be specified")
		}
		if *gp_listen_addr == "" && *gp_connect_to == "" {
			log.Fatal("One of --connect-to or --listen-addr options must be specified")
		}

		if *gp_scriptfile != "" {
			log.Warning("When started as a daemon, the --script option is ignored !")
		}
	}
}

// Program starts here
func main() {

	command_line()

	if gp_mode == standaloneMode {
		// Always creates a Hub for Accept Result in SpawnUsers
		log.Debugf("*gp_listen_addr=%s", *gp_listen_addr)
		hub = newHub()

		// Read the scenario from file
		data, err := ioutil.ReadFile(*gp_scriptfile)
		if err != nil {
			log.Fatal(err)
		}

		if !createPlaybook([]byte(data), &gp_playbook, &gp_actions) {
			log.Fatalf("Error while processing the Script File")
		}

		if gp_playbook.DataFeeder.Type == "csv" {
			feeder.Csv(gp_playbook.DataFeeder, path.Dir(*gp_scriptfile))
		} else if gp_playbook.DataFeeder.Type != "" {
			log.Fatalf("Unsupported feeder type: %s", gp_playbook.DataFeeder.Type)
		}

		reporter.SimulationStart = time.Now()

		var outputfile string
		var dir string
		if *gp_outputdir == "" {
			d, _ := os.Getwd()
			dir = d + "/results"
		} else {
			dir = *gp_outputdir
		}
		outputfile = dir + "/data." + *gp_outputtype
		if err := reporter.InitReport(*gp_outputtype); err != nil {
			log.Fatal(err)
		}
		reporter.OpenResultsFile(outputfile)

		spawnUsers(&gp_playbook, &gp_actions)

		log.Infof("Done in %v", time.Since(reporter.SimulationStart))
		log.Infof("Building reports, please wait...")
		reporter.CloseResultsFile()

		err = reporter.CloseReport(*gp_python_cmd, *gp_viewerfile, outputfile, dir)
		if err != nil {
			log.Error(err.Error())
		}

	}
	if gp_mode == daemonMode {
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
	if gp_mode == managerMode {
		log.Debugf("Start manager mode on this address: %s", *gp_manager_addr)
		manager.Start(gp_manager_addr)
	}
}

// Create a Playbook from the YAML data
func createPlaybook(data []byte, playbook *config.TestDef, actions *[]action.Action) bool {
	err := yaml.UnmarshalStrict([]byte(data), playbook)
	if err != nil {
		log.Fatalf("YAML error: %v", err)
	}
	log.Debug("Playbook:")
	log.Debug(playbook)

	if !config.ValidateTestDefinition(playbook) {
		return false
	}

	var isValid bool
	*actions, isValid = action.BuildActionList(playbook, path.Dir(*gp_scriptfile))
	if !isValid {
		return false
	}
	log.Debug("Tests Definition:")
	log.Debug(playbook)

	return true
}

// Launch VUs
func spawnUsers(playbook *config.TestDef, actions *[]action.Action) {
	resultsChannel := make(chan reporter.SampleReqResult, 10000)
	go reporter.AcceptResults(resultsChannel, &VU_count, &hub.broadcast)
	VU_start = time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < playbook.Users; i++ {
		wg.Add(1)
		VU_count++
		//UID := strconv.Itoa(rand.Intn(playbook.Users+1) + 10000)
		UID := strconv.Itoa(os.Getpid()*100000 + i)
		go launchActions(playbook, resultsChannel, &wg, actions, UID)
		waitDuration := float32(playbook.Rampup) / float32(playbook.Users)
		time.Sleep(time.Duration(int(1000*waitDuration)) * time.Millisecond)

		// In daemon mode, we may receive an order to stop the load test
		if gp_daemon_status == STOPPING_NOW {
			log.Info("Stop now")
			break
		}
	}
	log.Info("All users started, waiting at WaitGroup")
	wg.Wait()
	reporter.StopResults()
}

// Called once per each VU
func launchActions(playbook *config.TestDef, resultsChannel chan reporter.SampleReqResult, wg *sync.WaitGroup, actions *[]action.Action, UID string) {
	var sessionMap = make(map[string]string)

	i := 0
actionLoop:
	for (playbook.Iterations == -1) || (i < playbook.Iterations) {

		// Make sure the sessionMap is cleared before each iteration - except for the UID which stays
		cleanSessionMapAndResetUID(UID, sessionMap, playbook)

		// If we have feeder data, pop an item and push its key-value pairs into the sessionMap
		feedSession(playbook, sessionMap)

		// Iterate over the actions. Note the use of the command-pattern like Execute method on the Action interface
	iterLoop:
		for _, action := range *actions {
			if action != nil {
				//action.(Action).Execute(resultsChannel, sessionMap)
				if !action.Execute(resultsChannel, sessionMap, playbook) {
					// An error occurred : continue, stop the vu or stop the test ?
					switch playbook.OnError {
					case config.ERR_CONTINUE:
						log.Info("Continue on error")
						break
					case config.ERR_STOP_ITERATION:
						log.Info("Stop this iteration")
						break iterLoop
					case config.ERR_STOP_TEST:
						log.Info("Stop test on error")
						gp_daemon_status = STOPPING_NOW
						break actionLoop
					case config.ERR_STOP_VU:
						log.Info("Stop VU on error")
						break actionLoop
					}
				}
			}
		}
		if playbook.Iterations != -1 {
			i++
		} else {
			ti := time.Now()
			if ti.Sub(VU_start) > time.Duration(playbook.Duration)*time.Second {
				//log.Info("finished", time.Duration(t.Duration) * time.Second, ti.Sub(VU_start))
				break
			}
		}
	}
	wg.Done()
	VU_count--
}

func cleanSessionMapAndResetUID(UID string, sessionMap map[string]string, playbook *config.TestDef) {
	// Optimization? Delete all entries rather than reallocate map from scratch for each new iteration.
	for k := range sessionMap {
		delete(sessionMap, k)
	}

	// Set permanent variable and variables from playbook
	sessionMap["UID"] = UID
	sessionMap[config.HTTP_RESPONSE] = ""

	for k, v := range playbook.Variables {
		sessionMap[k] = v
	}
}

func feedSession(playbook *config.TestDef, sessionMap map[string]string) {
	if playbook.DataFeeder.Type != "" {
		go feeder.NextFromFeeder()       // Do async
		feedData := <-feeder.FeedChannel // Will block here until feeder delivers value over the FeedChannel
		for item := range feedData {
			sessionMap[item] = feedData[item]
		}
	}
}

// EOF
