package commands

import (
	"github.com/majeinfo/chaingun/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var appVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Displays the Player version",
	Long: `
Example usage:
player version`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Version mode started")
		log.Infof("Current version: %s", core.GitCommit)
	},
}

func init() {
	RootCmd.AddCommand(appVersionCmd)
}

