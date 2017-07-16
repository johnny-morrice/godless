package log

import (
	"fmt"
	"log"
	"strconv"

	"github.com/davecgh/go-spew/spew"
)

// Really basic logging helpers follow.

type LogLevel uint8

func (level LogLevel) String() string {
	switch level {
	case LOG_NOTHING:
		return "NOTHING"
	case LOG_DEBUG:
		return "DEBUG"
	case LOG_WARN:
		return "WARN"
	case LOG_ERROR:
		return "ERROR"
	case LOG_INFO:
		return "INFO"
	default:
		return strconv.Itoa(int(level))
	}
}

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
	Info("Log level set to %v", __LOG_LEVEL)
}

func CanLog(level LogLevel) bool {
	return __LOG_LEVEL >= level
}

func Debug(msg string, args ...interface{}) {
	if CanLog(LOG_DEBUG) {
		logMsg(LOG_DEBUG.String(), msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	if CanLog(LOG_INFO) {
		logMsg(LOG_INFO.String(), msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	if CanLog(LOG_WARN) {
		logMsg(LOG_WARN.String(), msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	if CanLog(LOG_ERROR) {
		logMsg(LOG_ERROR.String(), msg, args...)
	}
}

func logdie(msg string, args ...interface{}) {
	logMsg("FATAL", msg, args)
	panic(spew.Sprintf(msg, args))
}

func logMsg(level, msg string, args ...interface{}) {
	format := fmt.Sprintf("%s %s", level, msg)
	spewDump := spew.Sprintf(format, args...)
	log.Print(spewDump)
}
