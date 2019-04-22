package reporter

import (
    "fmt"
    "os/exec"
    "time"

    log "github.com/sirupsen/logrus"
)

// TODO: there should be an interface "outputter"...
const (
    jsonOutput = 0 + iota
    csvOutput
)

var (
    outputTypeMap = map[string]int{
        "json": jsonOutput,
        "csv":  csvOutput,
    }

    outputType      int
    SimulationStart time.Time
)

// InitReport initializes the report and memorizes the desired output type
func InitReport(otype string) error {
    if _, ok := outputTypeMap[otype]; !ok {
        return fmt.Errorf("Unsupported output type %s", otype)
    }

    SimulationStart = time.Now()
    outputType = outputTypeMap[otype]
    return nil
}

// CloseReport Build the final report
func CloseReport(pythonCmd string, viewerFile string, outputFile string, outputDir string) error {
    if outputType == csvOutput {
        // Build graphs
        log.Info("Launching Viewer")
        log.Infof("%s %s --data '%s' --output-dir '%s'",
            pythonCmd, viewerFile, outputFile, outputDir)
        cmd := exec.Command(pythonCmd, viewerFile,
            "--data", outputFile,
            "--output-dir", outputDir)
        err = cmd.Run()
        if err != nil {
            return fmt.Errorf("Viewer run failed: %s", err.Error())
        }
    }

    return nil
}

// EOF
