package config

import (
	log "github.com/sirupsen/logrus"
)

// Predefined Variables
const HTTP_RESPONSE = "HTTP_Response"
const MONGODB_LAST_INSERT_ID = "MONGODB_Last_Insert_ID"
const SQL_ROW_COUNT = "SQL_Row_Count"

const RE_FIRST = "first"
const RE_LAST = "last"
const RE_RANDOM = "random"

const ERR_CONTINUE = "continue"
const ERR_STOP_ITERATION = "stop_iteration"
const ERR_STOP_VU = "stop_vu"
const ERR_STOP_TEST = "stop_test"

const DFLT_TIMEOUT = 10
const DFLT_ERR = ERR_CONTINUE

type TestDef struct {
	Version        string                   `yaml:"version"`
	Iterations     int                      `yaml:"iterations"` // (mandatory) -1 implies use of "duration"
	Duration       int                      `yaml:"duration"`
	Users          int                      `yaml:"users"`
	Rampup         int                      `yaml:"rampup"`
	PersistentDBConn bool					`yaml:"persistent_db_connections"`	// (default is false)
	PersistentHttpSession bool 				`yaml:"persistent_http_sessions"` // (default is false)
	OnError        string                   `yaml:"on_error"` // continue (default) | stop_vu | stop_test
	HttpErrorCodes string                   `yaml:"http_error_codes"`
	GrpcProto	   string	                `yaml:"grpc_proto"`
	Timeout        int                      `yaml:"timeout"` // default is 10s
	DfltValues     Default                  `yaml:"default"`
	Variables      map[string]string        `yaml:"variables"`
	DataFeeder     Feeder                   `yaml:"feeder"`
	PreActions     []map[string]interface{} `yaml:"pre_actions"`
	Actions        []map[string]interface{} `yaml:"actions"`
}

type Default struct {
	Server     string `yaml:"server"` // Host or Host:Port
	Protocol   string `yaml:"protocol"`
	Method     string `yaml:"method"`
	Database   string `yaml:"database"`
	Collection string `yaml:"collection"`
	DBDriver   string `yaml:"db_driver"`
}

type Feeder struct {
	Type      string `yaml:"type"`
	Filename  string `yaml:"filename"`
	Separator string `yaml:"separator"`
}

type VUContext struct {
	InitObject	interface{}
	CloseFunc	func(*VUContext)
}

// Validate the Test Definition Consistency
func ValidateTestDefinition(t *TestDef) bool {
	valid := true
	if t.Version == "" {
		t.Version = "v1"
	}
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
	if t.Users <= 0 {
		log.Error("Users must be > 0")
		valid = false
	}
	if t.OnError == "" {
		t.OnError = DFLT_ERR
	} else {
		if t.OnError != ERR_CONTINUE && t.OnError != ERR_STOP_TEST && t.OnError != ERR_STOP_VU && t.OnError != ERR_STOP_ITERATION {
			log.Error("onerror parameter must be one of 'continue', 'stop_iteration', stop_vu' or 'stop_test'")
			valid = false
		}
	}

	if t.DfltValues.Method != "" && !IsValidHTTPMethod(t.DfltValues.Method) {
		log.Errorf("Default Http Action must specify a valid HTTP method: GET, POST, PUT, HEAD or DELETE: %s", t.DfltValues.Method)
		valid = false
	}

	if t.DfltValues.DBDriver != "" && !IsValidDBDriver(t.DfltValues.DBDriver) {
		log.Errorf("Default DB Driver must specify a valid driver (mysql): %s", t.DfltValues.DBDriver)
		valid = false
	}

	if t.GrpcProto != "" && t.DfltValues.Server == "" {
		log.Error("The Default Server name (and Port) is mandatory if grpc_proto is specified")
		valid = false
	}

	if t.Timeout == 0 {
		t.Timeout = DFLT_TIMEOUT
	}

	log.Infof("Playbook Version is %s", t.Version)

	return valid
}

// Check for method validity
func IsValidHTTPMethod(method string) bool {
	valid_methods := []string{"GET", "POST", "PUT", "HEAD", "DELETE"}

	return StringInSlice(method, valid_methods)
}

// Check for method validity
func IsValidMongoDBCommand(command string) bool {
	valid_commands := []string{"findone", "insertone", "deletemany", "drop"}

	return StringInSlice(command, valid_commands)
}

// Check for DBDriver validity
func IsValidDBDriver(db_driver string) bool {
	valid_drivers := []string{"mysql"}

	return StringInSlice(db_driver, valid_drivers)
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