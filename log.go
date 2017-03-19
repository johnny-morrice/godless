package godless

import (
	"fmt"
	"log"
)

// Really basic logging helpers follow.

func logdbg(msg string, args ...interface{}) {
	logMsg("DEBUG", msg, args...)
}

func logerr(msg string, args ...interface{}) {
	logMsg("ERROR", msg, args...)
}

func logwarn(msg string, args ...interface{}) {
	logMsg("WARN", msg, args...)
}

func logdie(msg string, args ...interface{}) {
	logMsg("FATAL", msg, args)
	panic(fmt.Sprintf(msg, args))
}

func logMsg(level, msg string, args ...interface{}) {
	log.Print(fmt.Sprintf(fmt.Sprintf("%v %v", level, msg), args...))
}
