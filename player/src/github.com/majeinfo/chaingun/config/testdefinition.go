package config

import (
    log "github.com/sirupsen/logrus"
)

const RE_FIRST = "first"
const RE_LAST = "last"
const RE_RANDOM = "random"

const ERR_CONTINUE = "continue"
const ERR_STOP_VU = "stop_vu"
const ERR_STOP_TEST = "stop_test"

type TestDef struct {
	Iterations int `yaml:"iterations"`					// (mandatory) -1 implies use of "duration"
	Duration int `yaml:"duration"`
	Users int `yaml:"users"`
	Rampup int `yaml:"rampup"`
	OnError string `yaml:"on_error"`					// continue (default) | stop_vu | stop_test
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
	if t.OnError == "" {
		t.OnError = ERR_CONTINUE
	} else {
		if t.OnError != ERR_CONTINUE && t.OnError != ERR_STOP_TEST && t.OnError != ERR_STOP_VU {
			log.Error("onerror parameter must be one of 'continue', 'stop_vu' or 'stop_test'")
			valid = false
		}
	}
    return valid
}

// EOF
