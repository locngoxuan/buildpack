package main

import (
	"flag"
	"fmt"
	"os"
	. "scm.wcs.fortna.com/lngo/buildpack"
)

const version = "0.1.0"

func main() {
	f := flag.NewFlagSet("buildpack", flag.ContinueOnError)
	f.Usage = func() {
		/**
		Do nothing
		 */
	}
	if len(os.Args) <= 1 {
		Usage(f)
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

	buildPack, err := NewBuildPack(action, f)
	if err != nil {
		LogFatal(BuildError{
			Action:  action,
			Phase:   "init",
			Err:     err,
			Message: "",
		})
	}
	result := Handle(buildPack)
	if result != nil {
		LogFatal(*result)
	}
	fmt.Println("[BUILDPACK] SUCCESS!!!")
	os.Exit(0)
}
