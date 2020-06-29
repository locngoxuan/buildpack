package main

import (
	"context"
	"fmt"
	"os"
	"scm.wcs.fortna.com/lngo/buildpack"
)

func printInfo(msg string, v ...interface{}) {
	if v != nil && len(v) > 0 {
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf(msg, v))
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, msg)
}

func printError(err error, msg string, v ...interface{}) {
	if v != nil && len(v) > 0 {
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf(msg, v), err)
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, msg, err)
}

func printFatal(err error, msg string, v ...interface{}) {
	printError(err, msg, v)
	os.Exit(1)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bp := buildpack.CreateBuildpack()
	bp.Run(ctx)
}
