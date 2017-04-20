package godless

import (
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
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
	panic(spew.Sprintf(msg, args))
}

func logMsg(level, msg string, args ...interface{}) {
	log.Print(spew.Sprintf(fmt.Sprintf("%v %v", level, msg), args...))
}
