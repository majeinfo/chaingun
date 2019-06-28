package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "net"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/majeinfo/chaingun/feeder"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// PlayerCommand describes the commands exchanged in JSON message
type PlayerCommand struct {
	Cmd      string `json:"cmd"`
	Value    string `json:"value"`
	MoreInfo string `json:"moreinfo"`
}

// DaemonStatus indicates the status of the Daemon
type DaemonStatus int

// PlayerStatus describes the structure of exchanged JSON message
type PlayerStatus struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Level  string `json:"level"`
	Msg    string `json:"msg"`
}

// PlayerResults describes the structure of exchanged Results !
type PlayerResults struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	Level      string `json:"level"`
	Msg        string `json:"msg"`
	HostName   string `json:"hostname"`
	ScriptFile string `json:"scriptfile"`
}

// Different states of remote daemon
const (
	IDLE DaemonStatus = 0 + iota
	READY_TO_RUN
	RUNNING
	STOPPING_NOW
)

var (
	gp_daemon_status DaemonStatus = IDLE
	statusString                  = []string{
		"Idle waiting for a Script",
		"Ready to run Script",
		"Running",
		"Stopping",
	}
	hub                 *Hub
	gp_outputfile       string
	g_results_available bool
	lock_status         sync.Mutex
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
	log.Debugf("Received Message: %s", msg)
	log.Debugf("Count of goroutines=%d", runtime.NumGoroutine())

	// Decode JSON message
	var cmd PlayerCommand
	err := json.Unmarshal(msg, &cmd)
	if err != nil {
		sendStatusError(c, "Message could not be decoded as JSON")
		return
	}

	switch cmd.Cmd {
	case "status":
		sendStatusOKMsg(c, "") //, statusString[gp_daemon_status])
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

func scriptCommand(c *Client, cmd *PlayerCommand) {
	// Check no run in progress
	if gp_daemon_status != IDLE && gp_daemon_status != READY_TO_RUN {
		sendStatusError(c, "Script ignored because daemon is not idle")
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
	if !createPlaybook(data, &gp_playbook, &gp_actions) {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while processing the Script data")
	} else {
		gp_valid_playbook = true
		gp_daemon_status = READY_TO_RUN
		sendStatusOKMsg(c, "Script received")

		// Ask for feeder data if needed
		if gp_playbook.DataFeeder.Type == "csv" {
			//feeder.Csv(gp_playbook.DataFeeder, path.Dir(*gp_scriptfile))
			log.Debugf("Ask for datafile %s", gp_playbook.DataFeeder.Filename)
			sendStatusOKMsg(c, fmt.Sprintf("Waiting for Data Feed: %s", gp_playbook.DataFeeder.Filename))
			sendGetDataFile(c, gp_playbook.DataFeeder.Filename)
		} else if gp_playbook.DataFeeder.Type != "" {
			sendStatusError(c, fmt.Sprintf("Unsupported feeder type: %s", gp_playbook.DataFeeder.Type))
		}
	}
}

func handleDataFeed(c *Client, cmd *PlayerCommand) {
	sendStatusOKMsg(c, "Data received")

	data, err := base64.StdEncoding.DecodeString(cmd.Value)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while decoding string from Base64")
		return
	}

	str_data := string(data[:])
	log.Debug(str_data)
	feeder.CsvInline(gp_playbook.DataFeeder, str_data)
}

func handleDataFile(c *Client, cmd *PlayerCommand) {
	sendStatusOKMsg(c, "File received")

	data, err := base64.StdEncoding.DecodeString(cmd.Value)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while decoding string from Base64")
		return
	}

	// Save the file locally
	err = ioutil.WriteFile(cmd.MoreInfo, data, 0644)
	if err != nil {
		gp_daemon_status = IDLE
		sendStatusError(c, "Error while writing file "+cmd.MoreInfo)
		return
	}
}

func getResultsCommand(c *Client, cmd *PlayerCommand) {
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
	var resp = &PlayerResults{
		Type:       "results",
		Status:     statusString[gp_daemon_status],
		Level:      "OK",
		Msg:        msg,
		HostName:   cmd.Value,
		ScriptFile: *gp_scriptfile,
	}
	j, _ := json.Marshal(resp)
	c.lock.Lock()
	defer c.lock.Unlock()
	c.conn.WriteMessage(websocket.TextMessage, j)

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
	var resp = &PlayerStatus{
		Type:   "status",
		Status: statusString[gp_daemon_status],
		Level:  level,
		Msg:    msg,
	}
	j, _ := json.Marshal(resp)
	c.lock.Lock()
	defer c.lock.Unlock()
	lock_status.Lock()
	defer lock_status.Unlock()
	c.conn.WriteMessage(websocket.TextMessage, j)
	log.Debug("exit sendStatus")
}

// Send a request to get data file content
func sendGetDataFile(c *Client, filename string) {
	var resp = &PlayerStatus{
		Type:   "getdata",
		Status: statusString[gp_daemon_status],
		Level:  "OK",
		Msg:    filename,
	}
	j, _ := json.Marshal(resp)
	c.lock.Lock()
	defer c.lock.Unlock()
	lock_status.Lock()
	defer lock_status.Unlock()
	c.conn.WriteMessage(websocket.TextMessage, j)
}

// EOF
