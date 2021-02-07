package main

import (
	"sync"
	"time"

	"github.com/majeinfo/chaingun/action"
	"github.com/majeinfo/chaingun/config"
)

const (
	standaloneMode = 0 + iota
	daemonMode
	managerMode
	batchMode
	graphOnlyMode
	designerMode
)

type playerFunc func()
type playerMode struct {
	mode int
	f    playerFunc
}

var (
	modeTypeMap = map[string]playerMode{
		"standalone": {standaloneMode, playStandaloneMode},
		"daemon":     {daemonMode, playDaemonMode},
		"manager":    {managerMode, playManagerMode},
		"batch":      {batchMode, playBatchMode},
		"graph-only": {graphOnlyMode, playGraphOnlyMode},
		"designer":   {designerMode, playDesignerMode},
	}
)

var (
	VU_start              time.Time
	VU_count              int
	lock_vu_count         sync.Mutex
	gp_mode               playerMode
	gp_valid_playbook     bool = false
	gp_listen_addr        *string
	gp_manager_addr       *string
	gp_repositorydir      *string
	gp_connect_to         *string
	gp_scriptfile         *string
	gp_outputdir          *string
	gp_outputtype         *string
	gp_injectors          *string
	gp_store_srv_resp_dir *string
	gp_no_log             *bool
	gp_display_srv_resp   *bool
	gp_trace              *bool
	gp_syntax_check_only  *bool
	gp_disable_dns_cache  *bool
	gp_trace_requests     *bool

	gp_playbook    config.TestDef
	gp_pre_actions []action.FullAction
	gp_actions     []action.FullAction

	GitCommit string = "1.1.5"

	gp_cpu_profile *string
	gp_mem_profile *string
)
