package logs

import (
	"github.com/sirupsen/logrus"
	"kikitoru/config"
)

//var L *slog.TextHandler

const queueLength = 32

type StructLog struct {
	MainLog  *Queue `json:"main_log"`
	Details  *Queue `json:"details"`
	Position int    `json:"position"`
	Total    int    `json:"total"`
	State    string `json:"state"`
}

var ScanLogs = StructLog{
	MainLog: NewQueue(1),
	Details: NewQueue(queueLength),
	State:   "running",
}

func InitLogger() {

	switch config.C.LogLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	case "":
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetReportCaller(true)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}
