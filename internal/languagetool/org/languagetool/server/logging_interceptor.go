package server

import (
	"regexp"
	"sync"
	"time"
)

// LoggingInterceptor ports org.languagetool.server.LoggingInterceptor (SQL timing log).
// Records statements without a real MyBatis stack.
type LoggingInterceptor struct {
	mu      sync.Mutex
	Entries []SQLLogEntry
}

// SQLLogEntry is one timed SQL execution record.
type SQLLogEntry struct {
	SQL       string
	Parameter any
	Duration  time.Duration
}

var sqlWhitespace = regexp.MustCompile(`\s+`)

func NewLoggingInterceptor() *LoggingInterceptor {
	return &LoggingInterceptor{}
}

// Intercept records duration of fn and collapses SQL whitespace.
func (l *LoggingInterceptor) Intercept(sql string, parameter any, fn func() error) error {
	start := time.Now()
	err := fn()
	dur := time.Since(start)
	if l == nil {
		return err
	}
	collapsed := sqlWhitespace.ReplaceAllString(sql, " ")
	l.mu.Lock()
	l.Entries = append(l.Entries, SQLLogEntry{SQL: collapsed, Parameter: parameter, Duration: dur})
	l.mu.Unlock()
	return err
}

func (l *LoggingInterceptor) Len() int {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.Entries)
}
