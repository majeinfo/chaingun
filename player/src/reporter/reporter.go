package reporter

import (
    "fmt"
    _ "os/exec"
    "time"

    "github.com/majeinfo/chaingun/viewer"
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
func CloseReport(outputFile, outputDir, scriptName string) error {
    if outputType == csvOutput {
        // Build graphs
        log.Info("Launching Viewer")
        err := viewer.BuildGraphs(outputFile, scriptName, outputDir)
        log.Infof("Graphs generated in directory %s", outputDir)
        return err
    }

    if outputType == jsonOutput {
        err := viewer.BuildJSON(outputFile, scriptName)
        return err
    }

    return nil
}

// EOF
