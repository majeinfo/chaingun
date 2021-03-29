package manager

// PlayerStatus describes the structure of exchanged JSON message
type PlayerStatus struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Level  string `json:"level"`
	Msg    string `json:"msg"`
	Detail string `json:"detail"`
}

// PlayerCommand describes the commands exchanged in JSON message
type PlayerCommand struct {
	Cmd      string `json:"cmd"`
	Value    string `json:"value"`
	MoreInfo string `json:"moreinfo"`
}

// DaemonStatus indicates the status of the Daemon
type DaemonStatus int

// PlayerResults describes the structure of exchanged Results !
type PlayerResults struct {
	Type       string `json:"type"`
	Status     string `json:"status"`
	Level      string `json:"level"`
	Msg        string `json:"msg"`
	HostName   string `json:"hostname"`
	ScriptFile string `json:"scriptfile"`
}
