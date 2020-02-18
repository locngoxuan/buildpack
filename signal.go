// +build !windows

package buildpack

import (
	"golang.org/x/sys/unix"
	"os"
)

func GetSingal() []os.Signal {
	return []os.Signal{
		os.Interrupt,
		os.Kill,
		unix.SIGHUP,
		unix.SIGCHLD,
		unix.SIGKILL,
		unix.SIGINT,
		unix.SIGTERM,
		unix.SIGQUIT,
	}
}