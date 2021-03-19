package main

import (
	"golang.org/x/sys/windows"
	"os"
)

func GetSingal() []os.Signal {
	return []os.Signal{
		os.Interrupt,
		os.Kill,
		windows.SIGHUP,
		windows.SIGKILL,
		windows.SIGINT,
		windows.SIGTERM,
		windows.SIGQUIT,
	}
}
