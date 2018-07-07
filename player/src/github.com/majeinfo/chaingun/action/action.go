package action

import (
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/config"
)

type Action interface {
	// Returns false if an error occurred during the execution
	Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool
}
