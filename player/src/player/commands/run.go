package commands

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"runtime"
	"runtime/pprof"
)

type rootFlags struct {
	cpu_profile	string
	mem_profile string
	verbose bool
	version bool
}

var (
	rootConfig rootFlags
	f_cpu_prof *os.File
)


// RootCmd defines the shell command usage for player.
var RootCmd = &cobra.Command{
	Use:   "player",
	Short: "A load-testing tool",
	Long: `This player uses Playbooks to stress Web servers.
It can be used in a "standalone" mode or in a "distributed" mode, it also embeds a "proxy" mode that can help you to create the Playbooks for HTTP/S protocols.
`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log_level := log.InfoLevel
		if rootConfig.verbose {
			log_level = log.DebugLevel
		}
		log.SetLevel(log_level)

		if rootConfig.cpu_profile != "" {
			var err error
			f_cpu_prof, err = os.Create(rootConfig.cpu_profile)
			if err != nil {
				log.Fatal("Could not create CPU profile: ", err)
			}
			//defer f.Close() // error handling omitted for example
			if err := pprof.StartCPUProfile(f_cpu_prof); err != nil {
				log.Fatal("Could not start CPU profile: ", err)
			}
			//defer pprof.StopCPUProfile()
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if rootConfig.cpu_profile != "" {
			f_cpu_prof.Close()
			pprof.StopCPUProfile()
		}

		if rootConfig.mem_profile != "" {
			f, err := os.Create(rootConfig.mem_profile)
			if err != nil {
				log.Fatal("Could not create memory profile: ", err)
			}
			defer f.Close() // error handling omitted for example
			runtime.GC()    // get up-to-date statistics
			if err := pprof.WriteHeapProfile(f); err != nil {
				log.Fatal("Could not write memory profile: ", err)
			}
			/*
				if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
					log.Fatal("Could not write memory profile: ", err)
				}
			*/
		}
	},
}

// Execute is a wrapper for the RootCmd.Execute method which will exit the program if there is an error.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&rootConfig.cpu_profile, "cpu-profile", "", "", "Write cpu profile to `file`")
	RootCmd.PersistentFlags().StringVarP(&rootConfig.mem_profile, "mem-profile", "", "", "Write memory profile to `file`")
	RootCmd.PersistentFlags().BoolVarP(&rootConfig.verbose, "verbose", "", false, "Set verbose mode")
	RootCmd.PersistentFlags().BoolVarP(&rootConfig.version, "version", "", false, "Displays the Version")
}
