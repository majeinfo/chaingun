package main

import (
	"flag"
	"os"

	"github.com/majeinfo/chaingun/action"
	log "github.com/sirupsen/logrus"
)

// Analyze the command line
func command_line() {
	mode := flag.String("mode", "standalone", "standalone|daemon|manager|batch|graph-only|designer")
	gp_listen_addr = flag.String("listen-addr", "127.0.0.1:12345", "Address and port to listen to (in daemon or designer mode)")
	gp_manager_addr = flag.String("manager-listen-addr", "127.0.0.1:8000", "Address and port to listen to (for the web interface in manager mode)")
	gp_repositorydir = flag.String("repository-dir", ".", "directory where to store results (in manager|batch mode)")
	gp_connect_to = flag.String("connect-to", "", "Address and port to connect to - in daemon mode (not supported yet)")
	verbose := flag.Bool("verbose", false, "Set verbose mode")
	version := flag.Bool("version", false, "Displays the Version")
	gp_scriptfile = flag.String("script", "", "Set the Script")
	gp_outputdir = flag.String("output-dir", "", "Set the output directory (standalone|graph-only mode)")
	gp_outputtype = flag.String("output-type", "csv", "Set the output type in file (csv|json)")
	gp_no_log = flag.Bool("no-log", false, "Disable the 'log' actions from the Script")
	gp_display_srv_resp = flag.Bool("display-response", false, "Used with verbose mode to display the Server Responses")
	gp_trace = flag.Bool("trace", false, "Generate a trace.out file useable by 'go tool trace' command (in standalone mode)")
	gp_syntax_check_only = flag.Bool("syntax-check-only", false, "Only validate the syntax of the Script")
	gp_disable_dns_cache = flag.Bool("disable-dns-cache", false, "Disable the embedded DNS cache which reduces the number of DNS requests")
	gp_trace_requests = flag.Bool("trace-requests", false, "Displays the HTTP/S requests and their return code")
	gp_injectors = flag.String("injectors", "", "Comma-separated list on already started injectors (ex: inject1:12345,inject2,inject3:1234) (manager|batch mode)")

	flag.Parse()

	if *version {
		log.Infof("Version: %s", GitCommit)
		os.Exit(0)
	}

	log_level := log.InfoLevel
	if *verbose {
		log_level = log.DebugLevel
	}
	log.SetLevel(log_level)

	// Check the mode
	var ok bool
	gp_mode, ok = modeTypeMap[*mode]
	if !ok {
		log.Fatalf("Unknown mode value: %s (allowed values are: standalone, daemon, manager, batch, designer or graph-only)", *mode)
	}
	log.Debugf("Player mode is %s", *mode)

	action.DisableLogAction(*gp_no_log)
	action.DisableDNSCache(*gp_disable_dns_cache)
	action.SetContext(gp_mode.mode == daemonMode, *gp_listen_addr, *gp_display_srv_resp, *gp_trace_requests)

	// Do some command line consistency tests
	if gp_mode.mode == standaloneMode {
		if *gp_scriptfile == "" {
			log.Fatal("When started in standalone mode, needs a script filename (option --script)")
		}
		checkNofileLimit()
	} else if gp_mode.mode == graphOnlyMode {
		// Use default parameters for outputdir and results
		if *gp_scriptfile == "" {
			log.Fatal("When started in graph-only mode, needs a script filename (option --script)")
		}
	} else if gp_mode.mode == daemonMode {
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
	} else if gp_mode.mode == batchMode {
		// Injectors and Script name are mandatory
		if *gp_injectors == "" {
			log.Fatal("When started in batch mode, needs a list of injectors (option --injectors)")
		}
		if *gp_scriptfile == "" {
			log.Fatal("When started in batch mode, needs a script filename (option --script)")
		}
	}
}
