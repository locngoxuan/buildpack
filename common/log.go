package common

import (
	"fmt"
	"io"
	"os"
	"scm.wcs.fortna.com/lngo/sqlbundle"
)

var logOutput io.Writer = os.Stdout

func SetLogOutput(w io.Writer) {
	logOutput = w
	sqlbundle.SetLogWriter(w)
}

func PrintInfo(msg string, v ...interface{}) {
	if v != nil && len(v) > 0 {
		_, _ = fmt.Fprintln(logOutput, fmt.Sprintf(msg, v...))
		return
	}
	_, _ = fmt.Fprintln(logOutput, msg)
}

func PrintErr(err error, msg string, v ...interface{}) {
	if v != nil && len(v) > 0 {
		_, _ = fmt.Fprintln(logOutput, fmt.Sprintf(msg, v...))
		_, _ = fmt.Fprintln(logOutput, fmt.Sprintf("reason: %v", err))
		return
	}
	_, _ = fmt.Fprintln(logOutput, msg)
	_, _ = fmt.Fprintln(logOutput, fmt.Sprintf("reason: %v", err))
}

func PrintFatal(err error, msg string, v ...interface{}) {
	PrintErr(err, msg, v...)
	os.Exit(1)
}
