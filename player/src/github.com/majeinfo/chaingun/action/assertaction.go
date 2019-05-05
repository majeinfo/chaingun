package action

import (
	"github.com/Knetic/govaluate"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

// AssertAction describes an assert Action
type AssertAction struct {
	Expression   string `yaml:"expression"`
	CompiledExpr *govaluate.EvaluableExpression
}

// Execute an assert Action
func (s AssertAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	result, err := utils.Evaluate(sessionMap, s.CompiledExpr, s.Expression)
	success := false

	if err == nil {
		// Check the result type and convert it into strings (float are converted into integer)
		switch result.(type) {
		case float64:
			success = result.(float64) != 0
		case string:
			success = result.(string) != ""
		case bool:
			success = result.(bool)
		default:
			log.Errorf("Error when evaluating expression: unknown type %v", result)
		}
	}

	if !success {
		log.Errorf("Assertion failed: %s", s.Expression)
		return false
	}

	log.Debugf("Assertion succeeded: %s", s.Expression)

	return true
}

// NewAssertAction creates a new Assert Action
func NewAssertAction(a map[interface{}]interface{}) (AssertAction, bool) {
	valid := true

	if a["expression"] == nil {
		log.Error("assert action needs 'expression' attribute")
		a["expression"] = ""
		valid = false
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(a["expression"].(string), getExpressionFunctions())

	assertAction := AssertAction{
		a["expression"].(string),
		expression,
	}

	if err != nil {
		log.Errorf("Expression '%s' cannot be compiled", err)
		valid = false
	}

	return assertAction, valid
}
