package action

import (
	"strconv"

	"github.com/Knetic/govaluate"
	"github.com/majeinfo/chaingun/config"
	"github.com/majeinfo/chaingun/reporter"
	log "github.com/sirupsen/logrus"
)

type SetVarAction struct {
	Variable     string `yaml:"variable"`
	Expression   string `yaml:"expression"`
	CompiledExpr *govaluate.EvaluableExpression
}

func (s SetVarAction) Execute(resultsChannel chan reporter.SampleReqResult, sessionMap map[string]string, playbook *config.TestDef) bool {
	// Create the variable if needed
	if _, err := sessionMap[s.Variable]; !err {
		log.Debugf("Variable ${%s} not set: creates it !", s.Variable)
		sessionMap[s.Variable] = ""
	}

	// Convert sessionMap into parameters for evaluation
	parameters := make(map[string]interface{}, len(sessionMap))
	for k, v := range sessionMap {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			parameters[k] = i
		} else {
			parameters[k] = v
		}
	}

	result, err := s.CompiledExpr.Evaluate(parameters)
	if err != nil {
		log.Errorf("Expression evaluation failed: %s", s.Expression)
		log.Errorf("%v", err)
	} else {
		// Check the result type and convert it into strings (float are converted into integer)
		switch result.(type) {
		case float64:
			sessionMap[s.Variable] = strconv.Itoa((int)(result.(float64)))
		case string:
			sessionMap[s.Variable] = result.(string)
		case bool:
			if result.(bool) {
				sessionMap[s.Variable] = "1"
			} else {
				sessionMap[s.Variable] = "0"
			}
		default:
			log.Errorf("Error when evaluating expression: unknown type %v", result)
		}
	}

	return true
}

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

	functions := map[string]govaluate.ExpressionFunction{
		"strlen": func(args ...interface{}) (interface{}, error) {
			length := len(args[0].(string))
			return (float64)(length), nil
		},
		"substr": func(args ...interface{}) (interface{}, error) {
			runes := []rune(args[0].(string))
			safeSubstring := string(runes[int(args[1].(float64)):int(args[2].(float64))])
			return safeSubstring, nil
		},
	}

	//expression, err := govaluate.NewEvaluableExpression(a["expression"].(string))
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(a["expression"].(string), functions)
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
