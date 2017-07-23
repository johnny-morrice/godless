package eval

import (
	"errors"
	"fmt"

	"github.com/johnny-morrice/godless/crdt"
)

type NamedFunction interface {
	FuncName() string
}

type MatchFunction interface {
	Match(literals []string, entries []crdt.Entry) bool
}

type NamedMatchFunction interface {
	NamedFunction
	MatchFunction
}

type NamedMatchFunctionLambda struct {
	Function MatchFunction
	Name     string
}

type MatchFunctionLambda func(literals []string, entries []crdt.Entry) bool

func (lambda MatchFunctionLambda) Match(literals []string, entries []crdt.Entry) bool {
	return lambda(literals, entries)
}

func (lambda NamedMatchFunctionLambda) Match(literals []string, entries []crdt.Entry) bool {
	return lambda.Function.Match(literals, entries)
}

func (lambda NamedMatchFunctionLambda) FuncName() string {
	return lambda.Name
}

type FunctionNamespace interface {
	GetFunction(functionName string) (NamedMatchFunction, error)
	PutFunction(function NamedMatchFunction) error
}

func MakeFunctionSet() FunctionNamespace {
	return sliceFunctionSet{}
}

type sliceFunctionSet struct {
	functions []NamedMatchFunction
}

func (set sliceFunctionSet) GetFunction(functionName string) (NamedMatchFunction, error) {
	for _, f := range set.functions {
		if f.FuncName() == functionName {
			return f, nil
		}
	}

	return nil, fmt.Errorf("No function for '%s", functionName)
}

func (set sliceFunctionSet) PutFunction(function NamedMatchFunction) error {
	_, err := set.GetFunction(function.FuncName())

	if err == nil {
		return fmt.Errorf("Duplicate function name: %s", function.FuncName())
	}

	set.functions = append(set.functions, function)
	return nil
}

func firstValue(literals []string, entries []crdt.Entry) (string, error) {
	var first string
	var found bool

	if len(literals) > 0 {
		first = literals[0]
		found = true
	} else {
		for _, entry := range entries {
			values := entry.GetValues()
			if len(values) > 0 {
				point := values[0]
				first = string(point.Text())
				found = true
				break
			}
		}
	}

	// No values: no firstValue.
	if !found {
		return "", errors.New("no firstValue")
	}

	return first, nil
}
