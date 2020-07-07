package common

import (
	"fmt"
	"io"
	"os"
)

func PrintLog(msg string, v ...interface{}) {
	PrintLogW(os.Stdout, msg, v)
}

func PrintLogW(w io.Writer, msg string, v ...interface{}) {
	if v != nil && len(v) > 0 {
		_, _ = fmt.Fprintln(w, fmt.Sprintf(msg, v...))
		return
	}
	_, _ = fmt.Fprintln(w, fmt.Sprintf(msg))
}
