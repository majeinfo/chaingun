package action

import (
	"strconv"

	"github.com/Knetic/govaluate"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

// AssertAction describes an assert Action
type AssertAction struct {
	Expression   string `yaml:"expression"`
	CompiledExpr *govaluate.EvaluableExpression
}

// Execute an assert Action
func (s AssertAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	// Convert sessionMap into parameters for evaluation
	parameters := make(map[string]interface{}, len(sessionMap))
	for k, v := range sessionMap {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			parameters[k] = i
		} else {
			parameters[k] = v
		}
	}

	success := false
	result, err := s.CompiledExpr.Evaluate(parameters)
	if err != nil {
		log.Errorf("Expression evaluation failed: %s", s.Expression)
		log.Errorf("%v", err)
	} else {
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
