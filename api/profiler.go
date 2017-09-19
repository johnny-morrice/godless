package api

import (
	"io"
	"sync"
)

type Profiler interface {
	NewTimer(name string) ProfileTimer
}

type ProfileCloser interface {
	Profiler
	io.Closer
}

type ProfileTimer interface {
	Start()
	Stop()
}

type TimerList struct {
	sync.Mutex
	timers []ProfileTimer
}

func (list *TimerList) StopAllTimers() {
	list.Lock()
	defer list.Unlock()

	for _, timer := range list.timers {
		if timer != nil {
			timer.Stop()
		}
	}
}

func (list *TimerList) StartTimer(timer ProfileTimer) {
	list.Lock()
	defer list.Unlock()

	written := list.writeOverNil(timer)
	if !written {
		list.timers = append(list.timers, timer)
	}

	timer.Start()
}

func (list *TimerList) StopTimer(timer ProfileTimer) {
	list.Lock()
	defer list.Unlock()

	timer.Stop()
	list.writeNil(timer)
}

func (list *TimerList) writeNil(timer ProfileTimer) {
	for i, other := range list.timers {
		if timer == other {
			list.timers[i] = nil
		}
	}
}

func (list *TimerList) writeOverNil(timer ProfileTimer) bool {
	for i, other := range list.timers {
		if other == nil {
			list.timers[i] = timer
			return true
		}
	}

	return false
}
