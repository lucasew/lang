package server

import (
	"sync"
	"time"
)

// RequestLimiter ports org.languagetool.server.RequestLimiter (token-bucket style).
type RequestLimiter struct {
	mu           sync.Mutex
	maxRequests  int
	window       time.Duration
	// ip → timestamps of recent requests
	byIP map[string][]time.Time
}

func NewRequestLimiter(maxRequests int, windowSeconds int) *RequestLimiter {
	if maxRequests <= 0 {
		maxRequests = 100
	}
	if windowSeconds <= 0 {
		windowSeconds = 60
	}
	return &RequestLimiter{
		maxRequests: maxRequests,
		window:      time.Duration(windowSeconds) * time.Second,
		byIP:        map[string][]time.Time{},
	}
}

// Allow records a request from ip and returns false if over limit.
func (l *RequestLimiter) Allow(ip string) bool {
	if l == nil {
		return true
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-l.window)
	times := l.byIP[ip]
	// prune
	i := 0
	for _, t := range times {
		if t.After(cutoff) {
			times[i] = t
			i++
		}
	}
	times = times[:i]
	if len(times) >= l.maxRequests {
		l.byIP[ip] = times
		return false
	}
	l.byIP[ip] = append(times, now)
	return true
}

// Count returns recent request count for ip.
func (l *RequestLimiter) Count(ip string) int {
	if l == nil {
		return 0
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-l.window)
	n := 0
	for _, t := range l.byIP[ip] {
		if t.After(cutoff) {
			n++
		}
	}
	return n
}
