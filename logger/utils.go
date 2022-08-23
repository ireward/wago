package logger

import (
	"path/filepath"
	"runtime"
	"strconv"
)

var (
	timeKey          = "ts"
	levelKey         = "level"
	nameKey          = "logger"
	messageKey       = "msg"
	stackTraceKey    = "stacktrace"
	loggingCallAtKey = "logging-call-at"
)

func caller(skip int) string {
	_, path, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}
	return filepath.Base(path) + ":" + strconv.Itoa(line)
}

func setDefaultMsg(msg string) string {
	if msg == "" {
		return "none"
	}
	return msg
}
