package commands

import (
	"github.com/majeinfo/chaingun/designer"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type designerFlags struct {
	listen_addr string
}

var designerConfig designerFlags

var appDesignerCmd = &cobra.Command{
	Use:   "design",
	Short: "Start a Web server application to help you write a Playbook",
	Long: `This mode starts a simple Web server that displays Forms
to help you create a new Playbook.

Example usage:
player designer`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("Start designer mode on this address: %s", designerConfig.listen_addr)
		designer.StartDesignerMode(designerConfig.listen_addr)
	},
}

func init() {
	RootCmd.AddCommand(appDesignerCmd)
	appDesignerCmd.Flags().StringVarP(&designerConfig.listen_addr, "listen-addr", "", "127.0.0.1:12345",
		"Address and port to listen to (ex: 127.0.0.1:8080)")
}

