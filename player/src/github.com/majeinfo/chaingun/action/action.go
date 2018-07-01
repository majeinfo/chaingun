package action

import (
    "github.com/majeinfo/chaingun/reporter"
)

type Action interface {
	Execute(resultsChannel chan reporter.HttpReqResult, sessionMap map[string]string)
}
