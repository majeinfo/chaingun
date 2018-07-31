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
	HttpErrorCodes string `yaml:"http_error_codes"`
	Timeout int `yaml:"timeout"`						// default is 10s
	DfltValues Default `yaml:"default"`
	DataFeeder Feeder `yaml:"feeder"`
	Actions []map[string]interface{} `yaml:"actions"`
}

type Default struct {
	Server string `yaml:"server"`						// Host or Host:Port
	Protocol string `yaml:"protocol"`
	Method string `yaml:"method"`
}

type Feeder struct {
	Type string `yaml:"type"`
	Filename string `yaml:"filename"`
	Separator string `yaml:"separator"`
}

// TODO: set default parms
// cookie_manager": { "enabled" :True, "clear_on_each_iteration": True }


// Validate the Test Definition Consistency
func ValidateTestDefinition(t *TestDef) (bool) {
    valid := true
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

    if t.DfltValues.Method != "" && !IsValidHTTPMethod(t.DfltValues.Method) {
        log.Error("Default Http Action must specify a valid HTTP method: GET, POST, PUT, HEAD or DELETE: %s", t.DfltValues.Method)
        valid = false
    }	

	if t.Timeout == 0 {
		t.Timeout = 10
	}
    return valid
}

// Check for method validity
func IsValidHTTPMethod(method string) bool {
	valid_methods := []string{"GET", "POST", "PUT", "HEAD", "DELETE"}

	return StringInSlice(method, valid_methods)
}

func StringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

// EOF
