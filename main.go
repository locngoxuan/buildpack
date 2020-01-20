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

action: init, snapshot, release, module
options:
	--m list of modules
	--v version
	--add apply for only module action
	--del apply for only module action
	--clean apply for snapshot and release
	--phase apply for snapshot and release
	--container run build command in container env
*/

func buildError(err BuildError) {
	if err.Err != nil {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] ERROR:", err.Action, err.Phase), err.Err)
	} else if len(strings.TrimSpace(err.Message)) > 0 {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] ERROR: %s", err.Action, err.Phase, err.Message))
	} else {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] UNKNOW ERROR", err.Action, err.Phase))
	}
	os.Exit(1)
}

func main() {
	f := flag.NewFlagSet("buildpack [init/snapshot/release] [OPTIONS]", flag.ContinueOnError)

	if len(os.Args) <= 1 {
		f.Usage()
		return
	}

	action := os.Args[1]
	err := verifyAction(action)
	if err != nil {
		buildError(BuildError{
			Action:  action,
			Phase:   BUILDPACK_PHASE_INIT,
			Err:     err,
			Message: "",
		})
	}

	buildPack := newBuildPack(action, f)
	result := buildPack.Handle()
	if result != nil {
		buildError(*result)
	}
	fmt.Println("[BUILDPACK] SUCCESS!!!")
	os.Exit(0)
}
