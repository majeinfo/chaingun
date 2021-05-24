package commands

import (
	_ "github.com/gorilla/websocket"
	"github.com/majeinfo/chaingun/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "net"
	_ "sync"
)

var injectConfig core.InjectStruct

var appInjectCmd = &cobra.Command{
	Use:   "inject",
	Short: "Starts the Player as a local Injector",
	Long: `This mode starts the Player as a local Injector: 
it plays the given Playbook (script) and generates the graph results.

Example usage:
player inject --script /path/to/playbook.yml`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Inject mode started")
		core.CheckNofileLimit()
		core.StartStandaloneMode(injectConfig)
	},
}

func init() {
	RootCmd.AddCommand(appInjectCmd)
	appInjectCmd.Flags().StringVarP(&injectConfig.Script, "script", "", "",
		"Set the Script filename")
	appInjectCmd.MarkFlagRequired("script")
	appInjectCmd.Flags().BoolVarP(&injectConfig.Trace, "trace", "", false, "Generate a trace.out file useable by 'go tool trace' command")
	appInjectCmd.Flags().BoolVarP(&injectConfig.Syntax_check_only, "syntax-check-only", "", false, "Only validate the syntax of the Script")
	appInjectCmd.Flags().StringVarP(&injectConfig.Listen_addr, "listen-addr", "", "127.0.0.1:12345",
		"Address and port to listen to (ex: 127.0.0.1:8080)")
	/*
	appInjectCmd.Flags().StringVarP(&injectConfig.Output_type, "output-type", "", "csv",
		"Set the output type in file (csv|json)")
	*/
	appInjectCmd.Flags().StringVarP(&injectConfig.Output_dir, "output-dir", "", "./results",
		"Set the output directory - where to put the data.csv file and the results")
	appInjectCmd.Flags().BoolVarP(&injectConfig.No_log, "no-log", "", false, "Disable the 'log' actions from the Script")
	appInjectCmd.Flags().BoolVarP(&injectConfig.Display_srv_resp, "display-response", "", false, "Used with verbose mode to display the Server Responses")
	appInjectCmd.Flags().StringVarP(&injectConfig.Store_srv_response_dir, "store-srv-response-dir", "", "",
		"Set the directory where to store the whole server response (often HTML)")
	appInjectCmd.Flags().BoolVarP(&injectConfig.Disable_dns_cache, "disable-dns-cache", "", false, "Disable the embedded DNS cache used to reduce the number of DNS requests")
	appInjectCmd.Flags().BoolVarP(&injectConfig.Trace_requests, "trace-requests", "", false, "Displays the requests and their return code")
}
