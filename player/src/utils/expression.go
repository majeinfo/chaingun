package utils

import (
	"math/rand"
	"strconv"
	"time"

	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
)

var (
	rand_src *rand.Rand
)

func init() {
	rand_src = rand.New(rand.NewSource(time.Now().UnixNano()))
}

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

// Implements embedded functions for compiled expressions
func GetExpressionFunctions() map[string]govaluate.ExpressionFunction {
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
		"random": func(args ...interface{}) (interface{}, error) {
			start := int(args[0].(float64))
			end := int(args[1].(float64))
			value := rand_src.Intn(end - start + 1) + start
			return (float64)(value), nil
		},
	}

	return functions
}

