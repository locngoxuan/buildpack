package common

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
		_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("reason: %v", err))
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, msg)
	_, _ = fmt.Fprintln(os.Stdout, fmt.Sprintf("reason: %v", err))
}

func PrintFatal(err error, msg string, v ...interface{}) {
	PrintErr(err, msg, v...)
	os.Exit(1)
}
