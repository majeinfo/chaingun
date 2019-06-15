package utils

import (
	"strconv"

	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
)

// Evaluate a precompiled expression
func Evaluate(sessionMap map[string]string, vulog *log.Entry, compiledExpr *govaluate.EvaluableExpression, expression string) (interface{}, error) {
	// Convert sessionMap into parameters for evaluation
	parameters := make(map[string]interface{}, len(sessionMap))
	for k, v := range sessionMap {
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			parameters[k] = i
		} else {
			parameters[k] = v
		}
	}

	result, err := compiledExpr.Evaluate(parameters)
	if err != nil {
		vulog.Errorf("Expression evaluation failed: %s", expression)
		vulog.Errorf("%v", err)
		return nil, err
	}

	vulog.Debugf("Expression evaluation succeeded: %s -> %s", expression, result)

	return result, nil
}
