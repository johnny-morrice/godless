package log

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
)

// Really basic logging helpers follow.

type LogLevel uint8

const (
	LOG_NOTHING = LogLevel(iota)
	LOG_ERROR
	LOG_WARN
	LOG_INFO
	LOG_DEBUG
)

var __LOG_LEVEL LogLevel

func SetLevel(level LogLevel) {
	__LOG_LEVEL = level
}

func CanLog(level LogLevel) bool {
	return __LOG_LEVEL >= level
}

func Debug(msg string, args ...interface{}) {
	if CanLog(LOG_DEBUG) {
		logMsg("DEBUG", msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	if CanLog(LOG_INFO) {
		logMsg("INFO", msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	if CanLog(LOG_WARN) {
		logMsg("WARN", msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	if CanLog(LOG_ERROR) {
		logMsg("ERROR", msg, args...)
	}
}

func logdie(msg string, args ...interface{}) {
	logMsg("FATAL", msg, args)
	panic(spew.Sprintf(msg, args))
}

func logMsg(level, msg string, args ...interface{}) {
	log.Print(spew.Sprintf(fmt.Sprintf("%v %v", level, msg), args...))
}
