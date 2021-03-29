package config

import (
	"bytes"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"testing"
)

type TestStructTrue struct {
	Name string
	Test TestDef
}

type TestStructFalse struct {
	Name string
	ExpectedMessage string
	Test TestDef
}

func TestConfigTrue(t *testing.T) {
	true_tests := []TestStructTrue{
		{ "OK001", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1 }},
		{ "OK002", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, OnError: "continue" }},
		{ "OK003", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, OnError: "stop_vu" }},
		{ "OK004", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, OnError: "stop_test" }},
		{ "OK005", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, DfltValues: Default{Method: "GET"}}},
		{ "OK006", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, DfltValues: Default{Method: "POST"}}},
		{ "OK007", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, DfltValues: Default{Method: "PUT"}}},
		{ "OK008", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, DfltValues: Default{Method: "HEAD"}}},
		{ "OK009", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, DfltValues: Default{Method: "DELETE"}}},
		{ "OK010", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, DfltValues: Default{DBDriver: "mysql"}}},
		{ "OK011", TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, GrpcProto: "dummy", DfltValues: Default{Server: "server"}}},
	}

	for _, test := range true_tests {
		if res := ValidateTestDefinition(&test.Test); !res {
			t.Errorf("Test failed: %s", test.Name)
		}
	}
}

func TestConfigFalse(t *testing.T) {
	false_tests := []TestStructFalse{
		{ "BAD001", "Iterations not set, must be > 0", TestDef{Iterations: -2, Duration: 1, Users: 1, Rampup: 1 }},
		{ "BAD002", "onerror parameter must be one of 'continue', 'stop_iteration', stop_vu' or 'stop_test'",
			TestDef{Iterations: 1, Duration: 1, Users: 1, Rampup: 1, OnError: "break" }},
		{ "BAD003", "When Iterations is -1, Duration must be set", TestDef{ Iterations: -1, Users: 1, Rampup: 1 }},
		{ "BAD004", "Rampup not defined. must be > -1", TestDef{ Iterations: -1, Duration: 1, Users: 1, Rampup: -1 }},
		{ "BAD005", "Users must be > 0", TestDef{ Iterations: -1, Duration: 1, Rampup: 1 }},
		{ "BAD006", "Default Http Action must specify a valid HTTP method: GET, POST, PUT, HEAD or DELETE",
			TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, DfltValues: Default{Method: "PATCH"}}},
		{ "BAD007", "Default DB Driver must specify a valid driver (mysql)",
			TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, DfltValues: Default{DBDriver: "oracle"}}},
		{ "BAD008", "The Default Server name (and Port) is mandatory if grpc_proto is specified",
			TestDef{ Iterations: 1, Duration: 1, Users: 1, Rampup: 1, GrpcProto: "dummy", DfltValues: Default{Server: ""}}},
	}

	for _, test := range false_tests {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		//t.Log(buf.String())

		if res := ValidateTestDefinition(&test.Test); res {
			t.Errorf("Test failed: %s", test.Name)
		}
		if !strings.Contains(buf.String(), test.ExpectedMessage ){
			t.Errorf("Wrong message: Expected: %s, got: %s", test.ExpectedMessage, buf.String())
		}
	}

	log.SetOutput(os.Stderr)
}
