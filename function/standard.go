package function

func StandardFunctions() FunctionNamespace {
	functions := MakeFunctionSet()
	err := functions.PutFunction(StrEq{})

	if err != nil {
		panic(err)
	}

	return functions
}
