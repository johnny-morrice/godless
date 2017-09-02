package log

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// Really basic logging helpers follow.

type LogLevel uint8

func Parse(text string) (LogLevel, error) {
	upperText := strings.ToUpper(text)
	switch upperText {
	case __DEBUG_TEXT:
		return DEBUG, nil
	case __INFO_TEXT:
		return INFO, nil
	case "WARNING":
		fallthrough
	case __WARN_TEXT:
		return WARN, nil
	case __ERROR_TEXT:
		return ERROR, nil
	case __NOTHING_TEXT:
		return NOTHING, nil
	}

	return DEBUG, fmt.Errorf("Unknown log level: '%s'", text)
}

func (level LogLevel) String() string {
	switch level {
	case NOTHING:
		return __NOTHING_TEXT
	case DEBUG:
		return __DEBUG_TEXT
	case WARN:
		return __WARN_TEXT
	case ERROR:
		return __ERROR_TEXT
	case INFO:
		return __INFO_TEXT
	default:
		return strconv.Itoa(int(level))
	}
}

const (
	NOTHING = LogLevel(iota)
	ERROR
	WARN
	INFO
	DEBUG
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
	if CanLog(DEBUG) {
		logMsg(DEBUG.String(), msg, args...)
	}
}

func Info(msg string, args ...interface{}) {
	if CanLog(INFO) {
		logMsg(INFO.String(), msg, args...)
	}
}

func Warn(msg string, args ...interface{}) {
	if CanLog(WARN) {
		logMsg(WARN.String(), msg, args...)
	}
}

func Error(msg string, args ...interface{}) {
	if CanLog(ERROR) {
		logMsg(ERROR.String(), msg, args...)
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

const __DEBUG_TEXT = "DEBUG"
const __INFO_TEXT = "INFO"
const __WARN_TEXT = "WARN"
const __ERROR_TEXT = "ERROR"
const __NOTHING_TEXT = "NOTHING"
