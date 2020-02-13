package main

import (
	"golang.org/x/sys/windows"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
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

func ForceClearOnTerminated(ch chan os.Signal, dirs ...string) {
	for {
		_ = <-ch
		signal.Stop(ch)
		for _, dir := range dirs {
			_ = os.RemoveAll(dir)
		}
	}
}

func Kill(pid int) error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(pid))
	kill.Stderr = os.Stderr
	kill.Stdout = os.Stdout
	return kill.Run()
}
