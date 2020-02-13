// +build !windows

package main

import (
	"golang.org/x/sys/unix"
	"os"
	"os/signal"
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

func ForceClearOnTerminated(ch chan os.Signal, handler HoolFunc) {
	for {
		_ = <-ch
		signal.Stop(ch)
		handler()
	}
}

func Kill(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil || p == nil {
		return err
	}
	err = p.Signal(unix.SIGINT)
	if err != nil {
		return err
	}
	_, _ = p.Wait()
	return err
}
