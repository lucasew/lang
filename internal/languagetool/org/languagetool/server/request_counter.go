package server

import "sync"

// RequestCounter ports org.languagetool.server.RequestCounter.
type RequestCounter struct {
	mu          sync.Mutex
	reqCount    int
	handleCount int
	handleIPs   map[int]string
}

func NewRequestCounter() *RequestCounter {
	return &RequestCounter{handleIPs: map[int]string{}}
}

func (c *RequestCounter) HandleCount() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.handleCount
}

func (c *RequestCounter) RequestCount() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.reqCount
}

func (c *RequestCounter) IncrementHandleCount(ip string, uniqueID int) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handleCount++
	c.handleIPs[uniqueID] = ip
}

func (c *RequestCounter) IncrementRequestCount() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reqCount++
	return c.reqCount
}

func (c *RequestCounter) DecrementHandleCount(uniqueID int) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handleCount--
	delete(c.handleIPs, uniqueID)
}

func (c *RequestCounter) DistinctIPs() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	seen := map[string]struct{}{}
	for _, ip := range c.handleIPs {
		seen[ip] = struct{}{}
	}
	return len(seen)
}
