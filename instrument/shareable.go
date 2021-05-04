package instrument

import (
	"log"
	"os"
	"runtime"
)

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
	BuildNumber   int
}

var extension string = ""

func init() {
	//detect os runtime
	if runtime.GOOS == "linux" {

	} else if runtime.GOOS == "windows" {
		extension = ".wins"
	} else if runtime.GOOS == "darwin" {
		extension = ".darwin"
	} else {
		log.Println("your local os either could not be recognized or is not supported")
		os.Exit(1)
	}
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
