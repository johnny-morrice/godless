// Package prof provides an event-based profiler for godless.
package prof

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/johnny-morrice/godless/api"
	"github.com/johnny-morrice/godless/log"
)

type WriterProfiler struct {
	sync.Mutex
	writer io.Writer
}

func MakeWriterProfiler(writer io.Writer) api.Profiler {
	return &WriterProfiler{writer: writer}
}

func (profiler *WriterProfiler) NewTimer(name string) api.ProfileTimer {
	return &writeProfileTimer{
		name:     name,
		profiler: profiler,
	}
}

func (profiler *WriterProfiler) writeEventDuration(name string, duration time.Duration) {
	profiler.Lock()
	defer profiler.Unlock()

	_, err := fmt.Fprintf(profiler.writer, "Event %s took %v\n", name, duration)

	if err != nil {
		log.Error("Profiler write error: %s", err.Error())
	}
}

type writeProfileTimer struct {
	name     string
	started  time.Time
	stopped  time.Time
	profiler *WriterProfiler
}

var zeroTime time.Time

func (timer *writeProfileTimer) Start() {
	if timer.stopped != zeroTime {
		panic("Cannot restart stopped timer")
	}

	timer.started = time.Now()
}

func (timer *writeProfileTimer) Stop() {
	if timer.started == zeroTime {
		panic("Timer stopped before start")
	}

	if timer.stopped != zeroTime {
		panic("Double stop of timer")
	}

	timer.stopped = time.Now()
	duration := timer.stopped.Sub(timer.started)
	timer.profiler.writeEventDuration(timer.name, duration)
}
