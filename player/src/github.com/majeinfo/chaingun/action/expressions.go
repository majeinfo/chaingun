package action

import (
	"github.com/Knetic/govaluate"
)

func getExpressionFunctions() map[string]govaluate.ExpressionFunction {
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

	return functions
}
