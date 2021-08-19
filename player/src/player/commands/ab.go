package commands


import (
	_ "github.com/gorilla/websocket"
	"github.com/majeinfo/chaingun/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	_ "net"
	_ "sync"
)

var abConfig core.AbStruct

var appAbCmd = &cobra.Command{
	Use:   "ab",
	Short: "Starts the Player as a local Injector in quick mode (like Apache ab)",
	Long: `This mode starts the Player as a local Injector: 
it plays the given request and generates the graph results.

Example usage:
player ab --request http://mysite.com/mypage.php`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Ab mode started")
		core.CheckNofileLimit()
		core.StartAbMode(abConfig)
	},
}

func init() {
	RootCmd.AddCommand(appAbCmd)
	appAbCmd.Flags().BoolVarP(&abConfig.Trace, "trace", "", false, "Generate a trace.out file useable by 'go tool trace' command")
	appAbCmd.Flags().StringVarP(&abConfig.Listen_addr, "listen-addr", "", "127.0.0.1:12345",
		"Address and port to listen to (ex: 127.0.0.1:8080)")
	appAbCmd.Flags().StringVarP(&abConfig.Output_dir, "output-dir", "", "./results",
		"Set the output directory - where to put the data.csv file and the results")
	appAbCmd.Flags().BoolVarP(&abConfig.No_log, "no-log", "", false, "Disable the 'log' actions from the Script")
	appAbCmd.Flags().BoolVarP(&abConfig.Display_srv_resp, "display-response", "", false, "Used with verbose mode to display the Server Responses")
	appAbCmd.Flags().StringVarP(&abConfig.Store_srv_response_dir, "store-srv-response-dir", "", "",
		"Set the directory where to store the whole server response (often HTML)")
	appAbCmd.Flags().BoolVarP(&abConfig.Disable_dns_cache, "disable-dns-cache", "", false, "Disable the embedded DNS cache used to reduce the number of DNS requests")
	appAbCmd.Flags().BoolVarP(&abConfig.Trace_requests, "trace-requests", "", false, "Displays the requests and their return code")
	appAbCmd.Flags().StringVarP(&abConfig.Request, "request", "", "", "URL of the request to be player")
	appAbCmd.MarkFlagRequired("request")
	appAbCmd.Flags().StringVarP(&abConfig.Method, "method", "", "GET", "HTTP method to use (GET=default, POST, PUT, HEAD)")
	appAbCmd.Flags().StringVarP(&abConfig.Body, "body", "", "", "Request body for POST requests")
	appAbCmd.Flags().IntVarP(&abConfig.Iterations, "iterations", "", -1,
		"Count of iterations for each VU (default value is -1 which means aonly the 'duration' parameter value is used")
	appAbCmd.Flags().IntVarP(&abConfig.Duration, "duration", "", 0,
		"Total duration (in seconds) of the stress - mandatory if 'iterations' is set to -1")
	appAbCmd.Flags().IntVarP(&abConfig.Users, "users", "", 0, "(mandatory) Count of VU to simulate")
	appAbCmd.MarkFlagRequired("users")
	appAbCmd.Flags().IntVarP(&abConfig.Rampup, "rampup", "", 0,
		"Gives the time in seconds that is use to launch the VU. New VUs are equally launched during this period. The default value is 0, so all VUs are launched immediatly")
	appAbCmd.MarkFlagRequired("users")
}
