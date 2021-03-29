package utils

import (
	_ "bytes"
	"github.com/Knetic/govaluate"
	log "github.com/sirupsen/logrus"
	_ "os"
	_ "strings"
	"testing"
)

type TestExpr struct {
	Name string
	Parms *map[string]string
	Expression string
}

var (
	parms1 = map[string]string{
		"size": "10",
	}
	parms2 = map[string]string{
		"name1": "alice",
		"name2": "bob",
	}
)

func TestGoodExpression(t *testing.T) {
	tests := []TestExpr{
		// Some Maths
		{ "OK001", &parms1, "3 + 4 == 7" },
		{ "OK002", &parms1, "size == 10" },
		{ "OK003", &parms1, "2*size > 19" },

		// Some Strings
		{ "OK050", &parms2, "(name1 + \" and \" + name2) == \"alice and bob\""},

		// Implemented Functions
		{ "OK080", &parms1, "random(10, 20) < 21 && random(10, 20) >= 10" },
		{ "OK081", &parms2, "strlen(name1) == 5" },
		{ "OK082", &parms2, "substr(name1,1,3) == \"li\""},
	}

	for _, test := range tests {
		success := false
		compiledExpr, err := govaluate.NewEvaluableExpressionWithFunctions(test.Expression, GetExpressionFunctions())
		if err != nil {
			t.Errorf("Test failed: %s (%s)", test.Name, err)
			continue
		}

		result, err := Evaluate(*test.Parms, log.WithFields(log.Fields{}), compiledExpr, test.Expression)
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
				t.Errorf("Error when evaluating expression: unknown type %v", result)
			}
		}

		if !success {
			t.Errorf("Error evaluating expression: %s (result=%v)", test.Expression, result)
		}
	}
}

func TestBadExpression(t *testing.T) {
	tests := []TestExpr{
		// Some Maths
		{ "BAD001", &parms1, "3 ++ 4 == 7" },

		// Some Strings
		{ "BAD050", &parms2, "(name1 + \" and \" * name2) == \"alice and bob\""},

		// Implemented Functions
		{ "BAD080", &parms2, "strings.ToUpper(name1) == \"ALICE\"" },
	}

	for _, test := range tests {
		compiledExpr, err := govaluate.NewEvaluableExpressionWithFunctions(test.Expression, GetExpressionFunctions())
		if err != nil {
			continue
		}

		result, err := Evaluate(*test.Parms, log.WithFields(log.Fields{}), compiledExpr, test.Expression)
		if err == nil {
			t.Errorf("Error not detected when evaluating expression: %s (result=%v)", test.Name, result)
		}
	}
}


