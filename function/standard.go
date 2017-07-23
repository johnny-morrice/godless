package function

func StandardFunctions() FunctionNamespace {
	functions := MakeFunctionNamespace()

	funcs := []NamedMatchFunction{
		StrEq{},
	}

	for _, f := range funcs {
		err := functions.PutFunction(f)

		if err != nil {
			panic("BUG: " + err.Error())
		}
	}

	return functions
}
