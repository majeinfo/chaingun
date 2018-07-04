package action

import (
    "github.com/majeinfo/chaingun/reporter"
)

type Action interface {
	// Returns false if an error occurred during the execution
	Execute(resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string) bool
}
