package reporter

import (
    "fmt"
    "time"
)

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

// EOF
