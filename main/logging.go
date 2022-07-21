package main

import (
	"log/syslog"

	"github.com/dumacp/go-logs/pkg/logs"
)

func newLog(logger *logs.Logger, prefix string, flags int, priority int) error {

	logg, err := syslog.NewLogger(syslog.Priority(priority), flags)
	if err != nil {
		return err
	}
	logger.SetLogError(logg)
	return nil
}

func initLogs(debug, logStd bool) {

	defer func() {
		if debug {
			return
		}
		logs.LogBuild.Disable()
	}()
	if logStd {
		return
	}
	newLog(logs.LogInfo, "[ info ] ", 0, 6)
	newLog(logs.LogWarn, "[ warn ] ", 0, 4)
	newLog(logs.LogError, "[ error ] ", 0, 3)
	newLog(logs.LogBuild, "[ build ] ", 0, 7)
}
