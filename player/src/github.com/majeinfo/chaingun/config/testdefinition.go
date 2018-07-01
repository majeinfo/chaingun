package config

import (
    log "github.com/sirupsen/logrus"
)

const FIRST = "first"
const LAST = "last"
const RANDOM = "random"

type TestDef struct {
	Iterations int `yaml:"iterations"`
	Duration int `yaml:"duration"`
	Users int `yaml:"users"`
	Rampup int `yaml:"rampup"`
	DataFeeder Feeder `yaml:"feeder"`
	Actions []map[string]interface{} `yaml:"actions"`
}

type Feeder struct {
	Type string `yaml:"type"`
	Filename string `yaml:"filename"`
	Separator string `yaml:"separator"`
}

// TODO: set default parms
// default_host, default_protocol, default_port, default_method
// on_error": "continue", # "stop_thread", "stop_test"
// cookie_manager": { "enabled" :True, "clear_on_each_iteration": True }



// Validate the Test Definition Consistency
func ValidateTestDefinition(t *TestDef) (bool) {
    var valid bool = true
    if t.Iterations <= 0 {
		if t.Iterations == -1 {
			if t.Duration < 1 {
				log.Error("When Iterations is -1, Duration must be set")
				valid = false
			}
		} else {
			log.Error("Iterations not set, must be > 0")
			valid = false
		}
    }
    if t.Rampup < 0 {
        log.Error("Rampup not defined. must be > -1")
        valid = false
    }
    if t.Users == 0 {
        log.Error("Users must be > 0")
        valid = false
    }
    return valid
}

// EOF
