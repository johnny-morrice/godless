package function

func StandardFunctions() FunctionNamespace {
	functions := MakeFunctionNamespace()

	funcs := []NamedMatchFunction{
		StrEq{},
		&StrGlob{},
		&StrRegexp{},
	}

	for _, f := range funcs {
		err := functions.PutFunction(f)

		if err != nil {
			panic("BUG: " + err.Error())
		}
	}

	return functions
}
