package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	_ "net"
	"net/http"
	"os"
	"path"
	"runtime"
	_ "sync"
	"time"

	_ "github.com/gorilla/websocket"
	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/feeder"
	"github.com/majeinfo/chaingun/manager"
	"github.com/majeinfo/chaingun/reporter"
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

// Start the WS Server
func startWsServer(listen_addr string) {
	//hub = newHub()
	go hub.run()

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
	log.Infof("Received Message: %s", msg[:int(math.Min(float64(len(msg)), 128))])
	log.Debugf("Count of goroutines=%d", runtime.NumGoroutine())

	// Decode JSON message
	var cmd manager.PlayerCommand
	err := json.Unmarshal(msg, &cmd)
	if err != nil {
		sendStatusError(c, "Message could not be decoded as JSON")
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
	case "datafeed":
		handleDataFeed(c, &cmd)
	case "datafile":
		handleDataFile(c, &cmd)
	case "get_results":
		getResultsCommand(c, &cmd)
	default:
		sendStatusError(c, fmt.Sprintf("Message not supported: %s", msg))
	}
	log.Debug("Message handled")
}

func preStartCommand(c *Client) {
	// Check no run in progress
	if gp_daemon_status != READY_TO_RUN {
		sendStatusError(c, "PreStart command ignored because daemon is not idle")
		return
	}

	// Check if there is a valid playbook
	if !gp_valid_playbook {
		sendStatusError(c, "PresStart command ignored because there is no valid playbook")
		return
	}

	gp_daemon_status = RUNNING
	sendStatusOK(c)
	go _preStartCommand(c)
}

func _preStartCommand(c *Client) {
	playPreActions(&gp_playbook, &gp_pre_actions)

	gp_daemon_status = READY_TO_RUN
	sendStatusOK(c)
}

func startCommand(c *Client) {
	// Check no run in progress
	if gp_daemon_status != READY_TO_RUN {
		sendStatusError(c, "Start command ignored because daemon is not idle")
		return
	}

	// Check if there is a valid playbook
	if !gp_valid_playbook {
		sendStatusError(c, "Start command ignored because there is no valid playbook")
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
		sendStatusError(c, fmt.Sprintf("Error while creating the temporary Results file: %s", err))
		return
	}

	g_results_available = false
	gp_outputfile = tmpfile.Name()
	log.Infof("Open outputfile: %s", gp_outputfile)
	if err := reporter.InitReport(*gp_outputtype); err != nil {
		sendStatusError(c, fmt.Sprintf("%s", err))
		return
	}
	reporter.OpenTempResultsFile(tmpfile)

	spawnUsers(&gp_playbook, &gp_actions)

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
		sendStatusError(c, "Stop command ignored because daemon is idle")
		return
	}

	gp_daemon_status = STOPPING_NOW
	sendStatusOK(c)
}

func scriptCommand(c *Client, cmd *manager.PlayerCommand) {
	// Check no run in progress
	if gp_daemon_status != IDLE && gp_daemon_status != READY_TO_RUN && gp_daemon_status != WAITING_FOR_FEEDER_DATA {
		sendStatusError(c, "Script ignored because daemon is not idle or ready to run or wating for data")
		return
	}
	log.Debugf("Original filename is %s", cmd.MoreInfo)
	gp_scriptfile = &cmd.MoreInfo

	// Receive a YAML Script
	gp_valid_playbook = false

	data, err := base64.StdEncoding.DecodeString(cmd.Value)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while decoding string from Base64")
		return
	}
	if !action.CreatePlaybook(gp_scriptfile, data, &gp_playbook, &gp_pre_actions, &gp_actions) {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while processing the Script data")
	} else {
		gp_valid_playbook = true
		gp_daemon_status = READY_TO_RUN
		sendStatusOKMsg(c, "Script received")
		log.Infof("Script received %s", cmd.MoreInfo)

		if gp_playbook.DataFeeder.Type == "csv" {
			if !feeder.Csv(gp_playbook.DataFeeder, path.Dir(*gp_scriptfile)) {
				sendStatusError(c, fmt.Sprintf("Could not load feeder data"))
			}
		} else if gp_playbook.DataFeeder.Type != "" {
			sendStatusError(c, fmt.Sprintf("Unsupported feeder type: %s", gp_playbook.DataFeeder.Type))
		}
	}
}

func handleDataFeed(c *Client, cmd *manager.PlayerCommand) {
	data, err := base64.StdEncoding.DecodeString(cmd.Value)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while decoding string from Base64")
		return
	}

	str_data := string(data[:])
	log.Debug(str_data)
	feeder.CsvInline(gp_playbook.DataFeeder, str_data)
	gp_daemon_status = READY_TO_RUN

	sendStatusOKMsg(c, "Data received")
}

func handleDataFile(c *Client, cmd *manager.PlayerCommand) {
	data, err := base64.StdEncoding.DecodeString(cmd.Value)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while decoding string from Base64")
		return
	}

	// Save the file locally - compute the location and creates directories
	//outputdir := path.Dir(path.Dir(*gp_scriptfile) + "/" + cmd.MoreInfo)
	outputdir := path.Dir("./" + cmd.MoreInfo)
	stat, err := os.Stat(outputdir)
	if os.IsNotExist(err) {
		log.Debugf("Must create the Output Directory %s", outputdir)
		if err := os.MkdirAll(outputdir, 0755); err != nil {
			sendStatusError(c, fmt.Sprintf("Cannot create Output Directory %s: %s", outputdir, err.Error()))
			return
		}
	} else if stat.Mode().IsRegular() {
		sendStatusError(c, fmt.Sprintf("Output Directory %s already exists as a file !", outputdir))
		return
	}

	err = ioutil.WriteFile(cmd.MoreInfo, data, 0644)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, fmt.Sprintf("Error while writing file %s", cmd.MoreInfo))
		return
	}

	sendStatusOKMsg(c, "File received")
	log.Infof("Received file %s", cmd.MoreInfo)
}

func getResultsCommand(c *Client, cmd *manager.PlayerCommand) {
	if !g_results_available {
		sendStatusError(c, "No Results available")
		return
	}

	sendStatusOKMsg(c, "Sending Results")

	data, err := ioutil.ReadFile(gp_outputfile)
	if err != nil {
		sendStatusError(c, "Error while readin File Results")
		return
	}

	msg := string(data)
	var resp = &manager.PlayerResults{
		Type:       "results",
		Status:     statusString[gp_daemon_status],
		Level:      "OK",
		Msg:        msg,
		HostName:   cmd.Value,
		ScriptFile: *gp_scriptfile,
	}
	j, _ := json.Marshal(resp)
	//c.send <- j
	sendToManager(c, j)

	sendStatusOKMsg(c, "Results sent successfully")
}

// Send OK
func sendStatusOK(c *Client) {
	sendStatus(c, "OK", "")
}

// Send OK
func sendStatusOKMsg(c *Client, msg string) {
	sendStatus(c, "OK", msg)
}

// Send an Error
func sendStatusError(c *Client, msg string) {
	log.Error(msg)
	sendStatus(c, "ERR", msg)
}

// Send a Status back to the manager
func sendStatus(c *Client, level string, msg string) {
	log.Debugf("sendStatus: %s", msg)
	var resp = &manager.PlayerStatus{
		Type:   "status",
		Status: statusString[gp_daemon_status],
		Level:  level,
		Msg:    msg,
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

// EOF
