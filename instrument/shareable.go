package instrument

type BaseProperties struct {
	WorkDir       string
	OutputDir     string
	ShareDataDir  string
	Version       string
	Release       bool
	Patch         bool
	ModuleName    string
	ModulePath    string
	ModuleOutputs []string
	LocalBuild    bool
}

type Response struct {
	Success  bool
	ErrStack string
	Err      error
}

func responseSuccess() Response {
	return Response{
		Success: true,
		Err:     nil,
	}
}

func responseError(err error) Response {
	return Response{
		Success: false,
		Err:     err,
	}
}
func responseErrorWithStack(err error, stack string) Response {
	return Response{
		Success:  false,
		ErrStack: stack,
		Err:      err,
	}
}
