package core

import (
	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/config"
	"sync"
	"time"
)

const (
	StandaloneMode = 0 + iota
	DaemonMode
	ManagerMode
	BatchMode
	GraphOnlyMode
	DesignerMode
	ProxyMode
)

var (
	VU_start              time.Time
	VU_count              int
	lock_vu_count         sync.Mutex
	g_emergency_stop     bool = false
	g_lock_emergency_stop   sync.Mutex
	g_valid_playbook     bool = false
	g_playbook    config.TestDef
	g_pre_actions []action.FullAction
	g_actions     []action.FullAction
	g_outputtype         string = "csv"
	g_scriptfile         string
)

