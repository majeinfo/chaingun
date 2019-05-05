package utils

import (
	"strconv"

	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
)

// Evaluate a precompiled expression
func Evaluate(sessionMap map[string]string, compiledExpr *govaluate.EvaluableExpression, expression string) (interface{}, error) {
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
		log.Errorf("Expression evaluation failed: %s", expression)
		log.Errorf("%v", err)
		return nil, err
	}

	log.Debugf("Expression evaluation succeeded: %s -> %s", expression, result)

	return result, nil
}
