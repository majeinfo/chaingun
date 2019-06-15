package action

import (
	"strconv"

	"github.com/Knetic/govaluate"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	"github.com/majeinfo/chaingun/utils"
	log "github.com/sirupsen/logrus"
)

// SetVarAction describes a setvar Action
type SetVarAction struct {
	Variable     string `yaml:"variable"`
	Expression   string `yaml:"expression"`
	CompiledExpr *govaluate.EvaluableExpression
}

// Execute a setvar Action
func (s SetVarAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, vulog *log.Entry, playbook *config.TestDef) bool {
	// Create the variable if needed
	if _, err := sessionMap[s.Variable]; !err {
		vulog.Debugf("Variable ${%s} not set: creates it !", s.Variable)
		sessionMap[s.Variable] = ""
	}

	result, err := utils.Evaluate(sessionMap, vulog, s.CompiledExpr, s.Expression)

	if err == nil {
		// Check the result type and convert it into strings (float are converted into integer)
		switch result.(type) {
		case float64:
			sessionMap[s.Variable] = strconv.Itoa((int)(result.(float64)))
			vulog.Debugf("setvar sets variable %s with value %s", s.Variable, sessionMap[s.Variable])
		case string:
			sessionMap[s.Variable] = result.(string)
			vulog.Debugf("setvar sets variable %s with value %s", s.Variable, sessionMap[s.Variable])
		case bool:
			if result.(bool) {
				sessionMap[s.Variable] = "1"
			} else {
				sessionMap[s.Variable] = "0"
			}
			vulog.Debugf("setvar sets variable %s with value %s", s.Variable, sessionMap[s.Variable])
		default:
			vulog.Errorf("Error when evaluating expression: unknown type %v", result)
		}
	}

	return true
}

// NewSetVarAction creates a new setvar Action
func NewSetVarAction(a map[interface{}]interface{}) (SetVarAction, bool) {
	valid := true
	if a["variable"] == nil {
		log.Error("setvar action needs 'variable' attribute")
		a["variable"] = ""
		valid = false
	}
	if a["expression"] == nil {
		log.Error("setvar action needs 'expression' attribute")
		a["expression"] = ""
		valid = false
	}

	expression, err := govaluate.NewEvaluableExpressionWithFunctions(a["expression"].(string), getExpressionFunctions())
	setVarAction := SetVarAction{
		a["variable"].(string),
		a["expression"].(string),
		expression,
	}

	if err != nil {
		log.Errorf("Expression '%s' cannot be compiled", err)
		valid = false
	}
	return setVarAction, valid
}
