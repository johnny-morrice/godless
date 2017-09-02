package log

import "testing"

func TestCanLog(t *testing.T) {
	table := map[LogLevel][]LogLevel{
		NOTHING: []LogLevel{},
		ERROR:   []LogLevel{ERROR},
		WARN:    []LogLevel{ERROR, WARN},
		INFO:    []LogLevel{ERROR, WARN, INFO},
		DEBUG:   []LogLevel{ERROR, WARN, INFO, DEBUG},
	}

	userLogLevels := []LogLevel{
		ERROR, WARN, INFO, DEBUG,
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
