package main

import (
	"flag"
	"io/ioutil"
	_ "math/rand"
	_ "net"
	"os"
	"path"
	"runtime"
	"runtime/trace"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/feeder"
	"github.com/majeinfo/chaingun/manager"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	standaloneMode = 0 + iota
	daemonMode
	managerMode
	batchMode
	graphOnlyMode
)

var (
	modeTypeMap = map[string]int{
		"standalone": standaloneMode,
		"daemon":     daemonMode,
		"manager":    managerMode,
		"batch":      batchMode,
		"graph-only": graphOnlyMode,
	}
)

var (
	VU_start             time.Time
	VU_count             int
	lock_vu_count        sync.Mutex
	gp_mode              int
	gp_valid_playbook    bool = false
	gp_listen_addr       *string
	gp_manager_addr      *string
	gp_repositorydir     *string
	gp_connect_to        *string
	gp_scriptfile        *string
	gp_outputdir         *string
	gp_outputtype        *string
	gp_injectors         *string
	gp_no_log            *bool
	gp_display_srv_resp  *bool
	gp_trace             *bool
	gp_syntax_check_only *bool
	gp_disable_dns_cache *bool

	gp_playbook config.TestDef
	gp_actions  []action.FullAction
)

// Analyze the command line
func command_line() {
	mode := flag.String("mode", "standalone", "standalone|daemon|manager|batch|graph-only")
	gp_listen_addr = flag.String("listen-addr", "127.0.0.1:12345", "Address and port to listen to (in daemon mode)")
	gp_manager_addr = flag.String("manager-listen-addr", "127.0.0.1:8000", "Address and port to listen to (for the web interface in manager mode)")
	gp_repositorydir = flag.String("repository-dir", ".", "directory where to store results (in manager|batch mode)")
	gp_connect_to = flag.String("connect-to", "", "Address and port to connect to - in daemon mode (not supported yet)")
	verbose := flag.Bool("verbose", false, "Set verbose mode")
	gp_scriptfile = flag.String("script", "", "Set the Script")
	gp_outputdir = flag.String("output-dir", "", "Set the output directory (standalone|graph-only mode)")
	gp_outputtype = flag.String("output-type", "csv", "Set the output type in file (csv|json)")
	gp_no_log = flag.Bool("no-log", false, "Disable the 'log' actions from the Script")
	gp_display_srv_resp = flag.Bool("display-response", false, "Used with verbose mode to display the Server Responses")
	gp_trace = flag.Bool("trace", false, "Generate a trace.out file useable by 'go tool trace' command (in standalone mode)")
	gp_syntax_check_only = flag.Bool("syntax-check-only", false, "Only validate the syntax of the Script")
	gp_disable_dns_cache = flag.Bool("disable-dns-cache", false, "Disable the embedded DNS cache which reduces the number of DNS requests")
	gp_injectors = flag.String("injectors", "", "Comma-separated list on already started injectors (ex: inject1:12345,inject2,inject3:1234) (manager|batch mode)")

	flag.Parse()

	log_level := log.InfoLevel
	if *verbose {
		log_level = log.DebugLevel
	}
	log.SetLevel(log_level)
	action.DisableLogAction(*gp_no_log)
	action.DisableDNSCache(*gp_disable_dns_cache)
	action.SetContext(*gp_display_srv_resp)

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
			log.Fatal("When started in standalone mode, needs a script filename (option --script)")
		}
		checkNofileLimit()
	} else if gp_mode == graphOnlyMode {
		// Use default parameters for outputdir and results
		if *gp_scriptfile == "" {
			log.Fatal("When started in graph-only mode, needs a script filename (option --script)")
		}
	} else if gp_mode == daemonMode {
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
		checkNofileLimit()
	} else if gp_mode == batchMode {
		// Injectors and Script name are mandatory
		if *gp_injectors == "" {
			log.Fatal("When started in batch mode, needs a list of injectors (option --injectors)")
		}
		if *gp_scriptfile == "" {
			log.Fatal("When started in batch mode, needs a script filename (option --script)")
		}
	}
}

func checkNofileLimit() {
	var rlim syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim); err != nil {
		log.Fatalf("syscall.Getrlimit() failed: %s", err)
	}
	log.Infof("Maximum number of open file descriptors: %d", rlim.Cur)
	if rlim.Cur < 4096 {
		log.Warning("You should increase this value to a higher value")
	}
}

// Program starts here
func main() {

	command_line()

	if gp_mode == standaloneMode {
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

		if !createPlaybook([]byte(data), &gp_playbook, &gp_actions) {
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
	} else if gp_mode == graphOnlyMode {
		// Just the graph production (TODO: does not work for merged data)
		outputfile, dir := computeOutputFilename()
		if err := reporter.InitReport(*gp_outputtype); err != nil {
			log.Fatal(err)
		}

		if err := reporter.CloseReport(outputfile, dir, *gp_scriptfile); err != nil {
			log.Error(err.Error())
		}
	} else if gp_mode == daemonMode {
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
	} else if gp_mode == managerMode {
		log.Infof("Start manager mode on this address: %s", *gp_manager_addr)
		manager.Start(gp_manager_addr, gp_repositorydir, gp_injectors)
	} else if gp_mode == batchMode {
		log.Debug("Batch mode started")
		manager.StartBatch(gp_manager_addr, gp_repositorydir, gp_injectors, gp_scriptfile)
	}
}

// Create a Playbook from the YAML data
func createPlaybook(data []byte, playbook *config.TestDef, actions *[]action.FullAction) bool {
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
func spawnUsers(playbook *config.TestDef, actions *[]action.FullAction) {
	resultsChannel := make(chan reporter.SampleReqResult, 10000)
	go reporter.AcceptResults(resultsChannel, &VU_count, &lock_vu_count, &hub.broadcast, gp_mode == daemonMode)
	VU_start = time.Now()
	wg := sync.WaitGroup{}
	for i := 0; i < playbook.Users; i++ {
		wg.Add(1)
		lock_vu_count.Lock()
		VU_count++
		lock_vu_count.Unlock()
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
func launchActions(playbook *config.TestDef, resultsChannel chan reporter.SampleReqResult, wg *sync.WaitGroup, actions *[]action.FullAction, UID string) {
	log.Debugf("launchActions called (%s)", UID)
	var sessionMap = make(map[string]string)

	i := 0
	vulog := log.WithFields(log.Fields{"vuid": UID, "iter": i, "action": ""})

actionLoop:
	for (playbook.Iterations == -1) || (i < playbook.Iterations) {
		vulog.Data["iter"] = i

		// Make sure the sessionMap is cleared before each iteration - except for the UID which stays
		cleanSessionMapAndResetUID(UID, sessionMap, playbook)

		// If we have feeder data, pop an item and push its key-value pairs into the sessionMap
		feedSession(playbook, sessionMap)

		// Iterate over the actions. Note the use of the command-pattern like Execute method on the Action interface
	iterLoop:
		for _, action := range *actions {
			if action.Action != nil {
				// Check for a "when" expression
				if action.CompiledWhen != nil {
					vulog.Debugf("Evaluate 'when' expression: %s", action.When)

					// if evaluation is False, skip the action
					result, err := utils.Evaluate(sessionMap, vulog, action.CompiledWhen, action.When)
					skip := false
					if err == nil {
						switch result.(type) {
						case float64:
							skip = result.(float64) == 0
						case string:
							skip = result.(string) == ""
						case bool:
							skip = !result.(bool)
						default:
							vulog.Errorf("Error when evaluating expression: unknown type %v", result)
						}
					}
					if skip {
						vulog.Infof("Action skipped due to its 'when' condition")
						continue
					}
				}
				if !action.Action.Execute(resultsChannel, sessionMap, vulog, playbook) {
					// An error occurred : continue, stop the vu or stop the test ?
					switch playbook.OnError {
					case config.ERR_CONTINUE:
						vulog.Info("Continue on error")
						break
					case config.ERR_STOP_ITERATION:
						vulog.Info("Stop this iteration")
						break iterLoop
					case config.ERR_STOP_TEST:
						vulog.Info("Stop test on error")
						gp_daemon_status = STOPPING_NOW
						break actionLoop
					case config.ERR_STOP_VU:
						vulog.Info("Stop VU on error")
						break actionLoop
					}
				}
			}
		}

		i++
		if playbook.Iterations == -1 {
			ti := time.Now()
			if ti.Sub(VU_start) > time.Duration(playbook.Duration)*time.Second {
				//log.Info("finished", time.Duration(t.Duration) * time.Second, ti.Sub(VU_start))
				break
			}
		}
	}
	wg.Done()
	lock_vu_count.Lock()
	VU_count--
	lock_vu_count.Unlock()
	log.Debugf("exit launchActions (%s)", UID)
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
