package godless

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
)

// Really basic logging helpers follow.

type LogLevel uint8

const (
	LOG_DEBUG = LogLevel(iota)
	LOG_INFO
	LOG_WARN
	LOG_ERROR
	LOG_NOTHING
)

var __LOG_LEVEL LogLevel

func SetDebugLevel(level LogLevel) {
	__LOG_LEVEL = level
}

func init() {
	SetDebugLevel(LOG_DEBUG)
}

func canLog(level LogLevel) bool {
	return __LOG_LEVEL <= level
}

func logdbg(msg string, args ...interface{}) {
	if canLog(LOG_DEBUG) {
		logMsg("DEBUG", msg, args...)
	}
}

func loginfo(msg string, args ...interface{}) {
	if canLog(LOG_INFO) {
		logMsg("INFO", msg, args...)
	}
}

func logwarn(msg string, args ...interface{}) {
	if canLog(LOG_WARN) {
		logMsg("WARN", msg, args...)
	}
}

func logerr(msg string, args ...interface{}) {
	if canLog(LOG_ERROR) {
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
