package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

/**

Usage:
buildpack [action] [options]

action: init, snapshot, release
options:
	--m list of modules
	--v version
*/

type BuildError struct {
	Err     error
	Action  string
	Phase   string
	Message string
}

func newError(phase, message string, err error) *BuildError {
	return &BuildError{
		Action:  runtimeParams.Action,
		Phase:   phase,
		Err:     err,
		Message: message,
	}
}

type ActionHandler func(f *flag.FlagSet) *BuildError

var actions map[string]ActionHandler
var runtimeParams BuildPackRuntimeParams

func buildError(err BuildError) {
	if err.Err != nil {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] ERROR:", err.Action, err.Phase), err)
	} else if len(strings.TrimSpace(err.Message)) > 0 {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] ERROR: %s", err.Action, err.Phase, err.Message))
	} else {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] UNKNOW ERROR", err.Action, err.Phase))
	}
	os.Exit(1)
}

const (
	ACTION_INIT     = "init"
	ACTION_SNAPSHOT = "snapshot"
	ACTION_RELEASE  = "release"

	BUILPACK_FILE = "buildpack.yml"
)

func init() {
	actions = make(map[string]ActionHandler)
	actions[ACTION_INIT] = ActionInitHandler
	actions[ACTION_SNAPSHOT] = ActionSnapshotHandler
	actions[ACTION_RELEASE] = ActionReleaseHandler

	runtimeParams = BuildPackRuntimeParams{}
}

func main() {
	f := flag.NewFlagSet("buildpack [init/snapshot/release] [OPTIONS]", flag.ContinueOnError)

	if len(os.Args) <= 1 {
		f.Usage()
		return
	}

	action := os.Args[1]

	err := f.Parse(os.Args[2:])
	if err != nil {
		buildError(BuildError{
			Action: action,
			Err:    err,
			Phase:  "init-config",
		})
	}

	runtimeParams.Action = action
	actionHandler, ok := actions[action]

	if !ok {
		buildError(BuildError{
			Action:  action,
			Phase:   "init-action",
			Message: "action not found",
		})
	}

	reult := actionHandler(f)
	if reult != nil {
		buildError(*reult)
	}
	fmt.Println("[BUILDPACK] SUCCESS!!!")
	os.Exit(0)
}
