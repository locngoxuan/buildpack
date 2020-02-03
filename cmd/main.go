package main

import (
	"flag"
	"fmt"
	"os"
	. "scm.wcs.fortna.com/lngo/buildpack"
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
func main() {
	if len(os.Args) <= 1 {
		f := flag.NewFlagSet("buildpack [init/verify/snapshot/release] [OPTIONS]", flag.ContinueOnError)
		f.Usage()
		return
	}

	action := os.Args[1]
	err := VerifyAction(action)
	if err != nil {
		LogFatal(BuildError{
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
	}

	f := flag.NewFlagSet(fmt.Sprintf("buildpack %s [OPTIONS]", action), flag.ContinueOnError)
	buildPack, err := NewBuildPack(action, f)
	if err != nil {
		LogFatal(BuildError{
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
	}
	result := buildPack.Handle()
	if result != nil {
		LogFatal(*result)
	}
	fmt.Println("[BUILDPACK] SUCCESS!!!")
	os.Exit(0)
}
