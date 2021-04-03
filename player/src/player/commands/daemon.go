package commands

import (
	_ "github.com/gorilla/websocket"
	"github.com/majeinfo/chaingun/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "net"
	_ "sync"
)

var daemonConfig core.DaemonStruct

var appDaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts the Player as a remote Injector",
	Long: `This mode starts the Player as a remote Injector: 
it listen on a port to commands sent by a Manager. It is used
to parallelize injections on a target to increase the global load.

Example usage:
player daemon --listen-addr 10.0.0.5:8080`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Daemon mode started")
		log.Debugf("listen_addr=%s", daemonConfig.Listen_addr)
		core.CheckNofileLimit()
		core.StartDaemonMode(daemonConfig)
	},
}

func init() {
	RootCmd.AddCommand(appDaemonCmd)
	appDaemonCmd.Flags().StringVarP(&daemonConfig.Listen_addr, "listen-addr", "", "127.0.0.1:12345",
		"Address and port to listen to (ex: 127.0.0.1:8080)")
	appDaemonCmd.Flags().BoolVarP(&daemonConfig.No_log, "no-log", "", false, "Disable the 'log' actions from the Script")
	appDaemonCmd.Flags().BoolVarP(&daemonConfig.Disable_dns_cache, "disable-dns-cache", "", false, "Disable the embedded DNS cache used to reduce the number of DNS requests")
	appDaemonCmd.Flags().BoolVarP(&daemonConfig.Trace_requests, "trace-requests", "", false, "Displays the requests and their return code")
}
