package core

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"runtime"
	"time"
	"github.com/majeinfo/chaingun/feeder"
	"github.com/majeinfo/chaingun/manager"
	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

// Different states of remote daemon
const (
	IDLE manager.DaemonStatus = 0 + iota
	WAITING_FOR_FEEDER_DATA
	READY_TO_RUN
	RUNNING
	STOPPING_NOW
)

type DaemonStruct struct {
	No_log bool
	Disable_dns_cache bool
	Trace_requests bool
	Listen_addr string
}

var (
	gp_daemon_status manager.DaemonStatus = IDLE
	statusString                          = []string{
		"Idle waiting for a Script",
		"Waiting for feeder data",
		"Ready to run Script",
		"Running",
		"Stopping",
	}
	hub                 *Hub
	gp_outputfile       string
	g_results_available bool
)

func StartDaemonMode(daemonParms DaemonStruct) {
	action.DisableLogAction(daemonParms.No_log)
	action.DisableDNSCache(daemonParms.Disable_dns_cache)
	action.SetContext(false, daemonParms.Listen_addr, false, daemonParms.Trace_requests, "")

	// Always creates a Hub for Accept Result in SpawnUsers
	hub = newHub()

	if daemonParms.Listen_addr != "" {
		log.Debugf("Create server listening on: %s", daemonParms.Listen_addr)
		startWsServer(daemonParms.Listen_addr)
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

// Start the WS Server
func startWsServer(listen_addr string) {
	//hub = newHub()
	go hub.Run()

	//http.HandleFunc("/", cmdHandler)
	http.HandleFunc("/upgrade", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(listen_addr, nil)
	if err != nil {
		log.Fatalf("Could not listen to %s: %s", listen_addr, err)
	}
}

// Handler of Request in daemon mode - called by Client ReadPump()
// TODO: manager and worker should exchange their versions
func cmdHandler(c *Client, msg []byte) {
	log.Infof("Received Message: %s ...", msg[:int(math.Min(float64(len(msg)), 128))])
	log.Debugf("Count of goroutines=%d", runtime.NumGoroutine())

	// Decode JSON message
	var cmd manager.PlayerCommand
	err := json.Unmarshal(msg, &cmd)
	if err != nil {
		sendStatusError(c, "Message could not be decoded as JSON", err.Error())
		return
	}

	switch cmd.Cmd {
	case "status":
		sendStatusOKMsg(c, "") //, statusString[gp_daemon_status])
	case "pre_start":
		preStartCommand(c)
	case "start":
		startCommand(c)
	case "stop":
		stopCommand(c)
	case "script":
		scriptCommand(c, &cmd)
	case "getmd5":
		handleGetMD5(c, &cmd)
	case "datafile":
		handleDataFile(c, &cmd)
	case "nextchunk":
		handleDataChunk(c, &cmd)
	case "get_results":
		getResultsCommand(c, &cmd)
	default:
		sendStatusError(c, fmt.Sprintf("Message not supported: %s ...", msg[:int(math.Min(float64(len(msg)), 128))]), "")
	}
	log.Debug("Message handled")
}

func preStartCommand(c *Client) {
	// Check no run in progress
	if gp_daemon_status != READY_TO_RUN {
		sendStatusError(c, "PreStart command ignored because daemon is not idle", "")
		return
	}

	// Check if there is a valid playbook
	if !g_valid_playbook {
		sendStatusError(c, "PresStart command ignored because there is no valid playbook", "")
		return
	}

	gp_daemon_status = RUNNING
	sendStatusOKMsg(c, "Launch pre-actions")
	go _preStartCommand(c)
}

func _preStartCommand(c *Client) {
	playPreActions(&g_playbook, &g_pre_actions)

	gp_daemon_status = READY_TO_RUN
	sendStatusOK(c)
}

func startCommand(c *Client) {
	// Check no run in progress
	if gp_daemon_status != READY_TO_RUN {
		sendStatusError(c, "Start command ignored because daemon is not idle", "")
		return
	}

	// Check if there is a valid playbook
	if !g_valid_playbook {
		sendStatusError(c, "Start command ignored because there is no valid playbook", "")
		return
	}

	gp_daemon_status = RUNNING
	sendStatusOK(c)
	go _startCommand(c)
}

func _startCommand(c *Client) {
	reporter.SimulationStart = time.Now()

	tmpfile, err := ioutil.TempFile("", "example")
	if err != nil {
		sendStatusError(c, "Error while creating the temporary Results file", err.Error())
		return
	}

	g_results_available = false
	gp_outputfile = tmpfile.Name()
	log.Infof("Open outputfile: %s", gp_outputfile)
	if err := reporter.InitReport(g_outputtype); err != nil {
		sendStatusError(c, "Reporter could not Init", err.Error())
		return
	}
	reporter.OpenTempResultsFile(tmpfile)

	spawnUsers(&g_playbook, &g_actions, DaemonMode)

	log.Infof("Done in %v", time.Since(reporter.SimulationStart))
	log.Infof("Close Results File...")
	reporter.CloseResultsFile()
	g_results_available = true

	gp_daemon_status = READY_TO_RUN
	sendStatusOK(c)
}

func stopCommand(c *Client) {
	// Check  run in progress
	if gp_daemon_status != RUNNING {
		sendStatusError(c, "Stop command ignored because daemon is idle", "")
		return
	}

	gp_daemon_status = STOPPING_NOW
	sendStatusOK(c)
}

func scriptCommand(c *Client, cmd *manager.PlayerCommand) {
	// Check no run in progress
	if gp_daemon_status != IDLE && gp_daemon_status != READY_TO_RUN && gp_daemon_status != WAITING_FOR_FEEDER_DATA {
		sendStatusError(c, "Script ignored because daemon is not idle or ready to run or wating for data", "")
		return
	}
	log.Debugf("Original filename is %s", cmd.MoreInfo)
	g_scriptfile = cmd.MoreInfo

	// Receive a YAML Script
	g_valid_playbook = false

	data, err := base64.StdEncoding.DecodeString(cmd.Value)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while decoding string from Base64", err.Error())
		return
	}
	if !action.CreatePlaybook(g_scriptfile, data, &g_playbook, &g_pre_actions, &g_actions) {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while processing the Script data", "")
	} else {
		g_valid_playbook = true
		gp_daemon_status = READY_TO_RUN
		sendStatusOKMsg(c, "Script received")
		log.Infof("Script received %s", cmd.MoreInfo)

		if g_playbook.DataFeeder.Type == "csv" {
			if !feeder.Csv(g_playbook.DataFeeder, path.Dir(g_scriptfile)) {
				sendStatusError(c, fmt.Sprintf("Could not load feeder data %s", g_playbook.DataFeeder), "")
			}
		} else if g_playbook.DataFeeder.Type != "" {
			sendStatusError(c, fmt.Sprintf("Unsupported feeder type: %s", g_playbook.DataFeeder.Type), "")
		}
	}
}

func handleGetMD5(c *Client, cmd *manager.PlayerCommand) {
	// Must compute and return the MD5 sum of the filename given in Value.
	// If the file does not exist, returns an error
	if md5sum, err := utils.Hash_file_md5(cmd.Value); err == nil {
		sendStatusOKMsg(c, md5sum)
		log.Infof("Received GetMD5 for file %s", cmd.Value)
		log.Infof("MD5 sum is %s", md5sum)
	} else {
		sendStatusError(c, fmt.Sprintf("Could not compute MD5sum of file %s", cmd.Value), err.Error())
	}
}

func handleDataFile(c *Client, cmd *manager.PlayerCommand) {
	data, err := base64.StdEncoding.DecodeString(cmd.Value)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while decoding string from Base64", err.Error())
		return
	}

	// Save the file locally - compute the location and creates directories
	//outputdir := path.Dir(path.Dir(*gp_scriptfile) + "/" + cmd.MoreInfo)
	outputdir := path.Dir("./" + cmd.MoreInfo)
	stat, err := os.Stat(outputdir)
	if os.IsNotExist(err) {
		log.Debugf("Must create the Output Directory %s", outputdir)
		if err := os.MkdirAll(outputdir, 0755); err != nil {
			sendStatusError(c, fmt.Sprintf("Cannot create Output Directory %s", outputdir), err.Error())
			return
		}
	} else if stat.Mode().IsRegular() {
		sendStatusError(c, fmt.Sprintf("Output Directory %s already exists as a file !", outputdir), "")
		return
	}

	err = ioutil.WriteFile(cmd.MoreInfo, data, 0644)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, fmt.Sprintf("Error while writing file %s", cmd.MoreInfo), err.Error())
		return
	}

	sendStatusOKMsg(c, "File received")
	log.Infof("Received file %s", cmd.MoreInfo)
}

func handleDataChunk(c *Client, cmd *manager.PlayerCommand) {
	data, err := base64.StdEncoding.DecodeString(cmd.Value)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while decoding string from Base64", err.Error())
		return
	}

	// The file should already exist and we must write the data in append mode
	//outputdir := path.Dir(path.Dir(*gp_scriptfile) + "/" + cmd.MoreInfo)
	outputdir := path.Dir("./" + cmd.MoreInfo)
	stat, err := os.Stat(outputdir)
	if os.IsNotExist(err) {
		log.Debugf("Must create the Output Directory %s", outputdir)
		if err := os.MkdirAll(outputdir, 0755); err != nil {
			sendStatusError(c, fmt.Sprintf("Cannot create Output Directory %s", outputdir), err.Error())
			return
		}
	} else if stat.Mode().IsRegular() {
		sendStatusError(c, fmt.Sprintf("Output Directory %s already exists as a file !", outputdir), "")
		return
	}

	f, err := os.OpenFile(cmd.MoreInfo, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		sendStatusError(c, fmt.Sprintf("Cannot open the file %s in append mode", cmd.MoreInfo), err.Error())
		return
	}
	if _, err := f.Write(data); err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, fmt.Sprintf("Error while writing file %s", cmd.MoreInfo), err.Error())
		return
	}
	if err := f.Close(); err != nil {
		sendStatusError(c, fmt.Sprintf("Error while closing file %s", cmd.MoreInfo), err.Error())
		return
	}

	sendStatusOKMsg(c, "Chunk of file received")
	log.Infof("Received file chunk %s", cmd.MoreInfo)
}

func getResultsCommand(c *Client, cmd *manager.PlayerCommand) {
	if !g_results_available {
		sendStatusError(c, "No Results available", "")
		return
	}

	sendStatusOKMsg(c, "Sending Results")

	data, err := ioutil.ReadFile(gp_outputfile)
	if err != nil {
		sendStatusError(c, "Error while readin File Results", "")
		return
	}

	msg := string(data)
	var resp = &manager.PlayerResults{
		Type:       "results",
		Status:     statusString[gp_daemon_status],
		Level:      "OK",
		Msg:        msg,
		HostName:   cmd.Value,
		ScriptFile: g_scriptfile,
	}
	j, _ := json.Marshal(resp)
	//c.send <- j
	sendToManager(c, j)

	sendStatusOKMsg(c, "Results sent successfully")
}

// Send OK
func sendStatusOK(c *Client) {
	sendStatus(c, "OK", "", "")
}

// Send OK
func sendStatusOKMsg(c *Client, msg string) {
	sendStatus(c, "OK", msg, "")
}

// Send an Error
func sendStatusError(c *Client, msg string, detail string) {
	log.Error(msg)
	sendStatus(c, "ERR", msg, detail)
}

// Send a Status back to the manager
func sendStatus(c *Client, level string, msg string, detail string) {
	log.Debugf("sendStatus: %s", msg)
	var resp = &manager.PlayerStatus{
		Type:   "status",
		Status: statusString[gp_daemon_status],
		Level:  level,
		Msg:    msg,
		Detail: detail,
	}
	j, _ := json.Marshal(resp)
	//c.send <- j
	sendToManager(c, j)
	log.Debug("exit sendStatus")
}

// Send a request to get data file content
func sendGetDataFile(c *Client, filename string) {
	log.Debugf("sendGetDataFile: %s", filename)
	var resp = &manager.PlayerStatus{
		Type:   "getdata",
		Status: statusString[gp_daemon_status],
		Level:  "OK",
		Msg:    filename,
	}
	j, _ := json.Marshal(resp)
	//c.send <- j
	sendToManager(c, j)
	log.Debug("exit sendGetDataFile")
}

// Send data but take care on writing on closed channel !
func sendToManager(c *Client, data []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	c.send <- data
	return
}
