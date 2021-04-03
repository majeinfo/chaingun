package commands

import (
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type graphFlags struct {
	output_dir string
	output_type string
	script string
}

var graphConfig graphFlags

var appGraphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Starts the graph-only mode to regenerate the graphs from previous results",
	Long: `This mode starts the "graph-only" mode that can recreate the graphs
from a script filename and previously computed results (the data.csv file).

Example usage:
player graph --script /path/to/playbook.yml --output-dir /path/to/results`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Debug("Graph mode started")
		playGraphOnlyMode()
	},
}

func init() {
	RootCmd.AddCommand(appGraphCmd)
	appGraphCmd.Flags().StringVarP(&graphConfig.output_dir, "output-dir", "", "./results",
		"Set the output directory - must contain the data.csv file")
	appGraphCmd.Flags().StringVarP(&graphConfig.output_type, "output-type", "", "csv",
		"Set the output type in file (csv|json)")
	appGraphCmd.Flags().StringVarP(&graphConfig.script, "script", "", "",
		"Set the Script filename")
	appGraphCmd.MarkFlagRequired("script")
}

func playGraphOnlyMode() {
	// Just the graph production
	outputfile, dir := utils.ComputeOutputFilename(graphConfig.output_dir, graphConfig.output_type)
	if err := reporter.InitReport(graphConfig.output_type); err != nil {
		log.Fatal(err)
	}

	if err := reporter.CloseReport(outputfile, dir, graphConfig.script); err != nil {
		log.Error(err.Error())
	}
}


