package action

import (
	"github.com/Knetic/govaluate"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
)

// Action is an interface which is able to execute a Request
type Action interface {
	// Returns false if an error occurred during the execution
	Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool
}

// FullAction embeds the global parameters for all actions as well as an Action
type FullAction struct {
	When         string `yaml:"when"`
	CompiledWhen *govaluate.EvaluableExpression
	Action       Action
}
