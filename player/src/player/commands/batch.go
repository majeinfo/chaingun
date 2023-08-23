package commands

import (
	"github.com/majeinfo/chaingun/manager"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type batchFlags struct {
	script         string
	repository_dir string
	injectors      string
}

var batchConfig batchFlags

var appBatchCmd = &cobra.Command{
	Use:   "batch",
	Short: "Runs a Playbook using Injectors",
	Long: `This mode plays the given script using the specified Injectors.

Example usage:
player batch --injectors injector1:8000,injector2:9000 --script /path/to/script.yml`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Batch mode started")
		manager.StartBatch(batchConfig.repository_dir, batchConfig.injectors, batchConfig.script)
	},
}

func init() {
	RootCmd.AddCommand(appBatchCmd)
	appBatchCmd.Flags().StringVarP(&batchConfig.script, "script", "", "",
		"Set the Script filename")
	appBatchCmd.Flags().StringVarP(&batchConfig.repository_dir, "repository-dir", "", ".",
		"Directory where to store results")
	appBatchCmd.Flags().StringVarP(&batchConfig.injectors, "injectors", "", "",
		"Comma-separated list of already started injectors (ex: inject1:12345,inject2,inject3:1234)")
	appBatchCmd.MarkFlagRequired("script")
	appBatchCmd.MarkFlagRequired("injectors")
}
