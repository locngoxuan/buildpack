package instrument

type BaseProperties struct {
	WorkDir       string
	OutputDir     string
	ShareDataDir  string
	Version       string
	DevMode       bool
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

func ResponseSuccess() Response {
	return Response{
		Success: true,
		Err:     nil,
	}
}

func ResponseError(err error) Response {
	return Response{
		Success: false,
		Err:     err,
	}
}
func ResponseErrorWithStack(err error, stack string) Response {
	return Response{
		Success:  false,
		ErrStack: stack,
		Err:      err,
	}
}
