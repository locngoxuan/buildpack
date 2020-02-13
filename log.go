package buildpack

import (
	"fmt"
	"os"
	"strings"
)

func LogOnlyMsg(msg string) {
	fmt.Println(fmt.Sprintf("[INFO] %s", msg))
}

func LogInfo(bp BuildPack, msg string) {
	fmt.Println(fmt.Sprintf("[INFO] [%s:%s] %s", bp.Action, bp.Phase, msg))
}

func LogInfoWithoutPhase(bp BuildPack, msg string) {
	fmt.Println(fmt.Sprintf("[INFO] [%s] %s", bp.Action, msg))
}

func LogVerbose(bp BuildPack, msg string) {
	if !bp.RuntimeConfig.Verbose() {
		return
	}
	fmt.Println(fmt.Sprintf("[INFO] [%s:%s] %s", bp.Action, bp.Phase, msg))
}

func LogFatal(err BuildResult) {
	if err.Err != nil {
		fmt.Println(fmt.Sprintf("[FATAL] [%s:%s] :", err.Action, err.Phase), err.Err)
	} else if len(strings.TrimSpace(err.Message)) > 0 {
		fmt.Println(fmt.Sprintf("[FATAL] [%s:%s] : %s", err.Action, err.Phase, err.Message))
	} else {
		fmt.Println(fmt.Sprintf("[FATAL] [%s:%s] unknow error", err.Action, err.Phase))
	}
	os.Exit(1)
}
