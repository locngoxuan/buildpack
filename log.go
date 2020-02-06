package buildpack

import (
	"fmt"
	"os"
	"strings"
)

func LogOnlyMsg(msg string) {
	fmt.Println(fmt.Sprintf("[BUILDPACK] %s", msg))
}

func LogInfo(bp BuildPack, msg string) {
	fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] %s", bp.Action, bp.Phase, msg))
}

func LogInfoWithoutPhase(bp BuildPack, msg string) {
	fmt.Println(fmt.Sprintf("[BUILDPACK] [%s] %s", bp.Action, msg))
}

func LogVerbose(bp BuildPack, msg string) {
	if !bp.RuntimeConfig.Verbose() {
		return
	}
	fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] %s", bp.Action, bp.Phase, msg))
}

func LogFatal(err BuildResult) {
	if err.Err != nil {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] ERROR:", err.Action, err.Phase), err.Err)
	} else if len(strings.TrimSpace(err.Message)) > 0 {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] ERROR: %s", err.Action, err.Phase, err.Message))
	} else {
		fmt.Println(fmt.Sprintf("[BUILDPACK] [%s:%s] UNKNOW ERROR", err.Action, err.Phase))
	}
	os.Exit(1)
}
