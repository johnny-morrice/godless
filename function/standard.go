package function

func StandardFunctions() FunctionNamespace {
	functions := MakeFunctionNamespace()
	err := functions.PutFunction(StrEq{})

	if err != nil {
		panic(err)
	}

	return functions
}
