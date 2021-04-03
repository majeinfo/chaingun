package commands

import (
	"github.com/majeinfo/chaingun/manager"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type managerFlags struct {
	manager_listen_addr string
	repository_dir string
	injectors string
}

var managerConfig managerFlags

var appManagerCmd = &cobra.Command{
	Use:   "manage",
	Short: "Starts a Web Server that allows you to manage remote Injectors",
	Long: `This mode starts a simple Web Server that displays an interface
you can use to connect to remote Injectors (i.e player is run with "injector" mode).
You can send them Playbooks and start them in parallel to increase the load on
the target. Then the results are downloaded and merged locally.

Example usage:
player manager`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Start manager mode on this address: %s", managerConfig.manager_listen_addr)
		manager.Start(managerConfig.manager_listen_addr, managerConfig.repository_dir, managerConfig.injectors)
	},
}

func init() {
	RootCmd.AddCommand(appManagerCmd)
	appManagerCmd.Flags().StringVarP(&managerConfig.manager_listen_addr, "listen-addr", "", "127.0.0.1:8000",
		"Address and port to listen to (for the web interface in manager mode)")
	appManagerCmd.Flags().StringVarP(&managerConfig.repository_dir, "repository-dir", "", ".",
		"Directory where to store results (must be a relative directory)")
	appManagerCmd.Flags().StringVarP(&managerConfig.injectors, "injectors", "", "",
		"Comma-separated list of already started injectors (ex: inject1:12345,inject2,inject3:1234)")
}
