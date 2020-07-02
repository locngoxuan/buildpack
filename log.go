package buildpack

import (
	"fmt"
	"os"
)

func PrintInfo(msg string, v ...interface{}) {
	if v != nil && len(v) > 0 {
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf(msg, v...))
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, msg)
}

func PrintErr(err error, msg string, v ...interface{}) {
	if v != nil && len(v) > 0 {
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf(msg, v...))
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("err: %v", err))
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, msg, err)
}

func PrintFatal(err error, msg string, v ...interface{}) {
	PrintErr(err, msg, v...)
	os.Exit(1)
}
