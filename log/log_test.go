package log

import "testing"

func TestCanLog(t *testing.T) {
	table := map[LogLevel][]LogLevel{
		LOG_NOTHING: []LogLevel{},
		LOG_ERROR:   []LogLevel{LOG_ERROR},
		LOG_WARN:    []LogLevel{LOG_ERROR, LOG_WARN},
		LOG_INFO:    []LogLevel{LOG_ERROR, LOG_WARN, LOG_INFO},
		LOG_DEBUG:   []LogLevel{LOG_ERROR, LOG_WARN, LOG_INFO, LOG_DEBUG},
	}

	userLogLevels := []LogLevel{
		LOG_ERROR, LOG_WARN, LOG_INFO, LOG_DEBUG,
	}

	for current, permitted := range table {
		SetLevel(current)

		for _, level := range userLogLevels {
			loggable := CanLog(level)
			found := findLevel(level, permitted)

			if loggable != found {
				t.Errorf("Bad log permission '%v' at %v for %v", loggable, current, level)
			}
		}
	}
}

func findLevel(level LogLevel, levels []LogLevel) bool {
	for _, l := range levels {
		if level == l {
			return true
		}
	}

	return false
}
