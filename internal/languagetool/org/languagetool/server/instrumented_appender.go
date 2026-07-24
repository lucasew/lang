package server

import "sync"

// InstrumentedAppender ports org.languagetool.server.InstrumentedAppender counters.
// Prometheus registration deferred; in-process counters by level/logger.
const InstrumentedAppenderCounterName = "languagetool_logback_appender_total"

type InstrumentedAppender struct {
	mu       sync.Mutex
	counters map[string]int64 // key: level|logger
}

func NewInstrumentedAppender() *InstrumentedAppender {
	return &InstrumentedAppender{counters: map[string]int64{}}
}

// Append increments the counter for a log event.
func (a *InstrumentedAppender) Append(level, logger, marker, exception string) {
	if a == nil {
		return
	}
	key := level + "|" + logger
	a.mu.Lock()
	a.counters[key]++
	a.mu.Unlock()
	_ = marker
	_ = exception
}

func (a *InstrumentedAppender) Count(level, logger string) int64 {
	if a == nil {
		return 0
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.counters[level+"|"+logger]
}
