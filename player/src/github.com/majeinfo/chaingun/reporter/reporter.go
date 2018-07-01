package reporter

import (
    "time"
    //log "github.com/sirupsen/logrus"
)

var (
    output_type string
    SimulationStart time.Time
)

func InitReport(otype string) {
    SimulationStart = time.Now()
    output_type = otype
}

// EOF
