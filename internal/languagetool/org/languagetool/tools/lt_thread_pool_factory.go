package tools

import (
	"sync"
	"time"
)

// Well-known pool names from LtThreadPoolFactory.
const (
	ServerPoolName            = "lt-server-thread"
	TextCheckerPoolName       = "lt-text-checker-thread"
	RemoteRuleExecutingPool   = "remote-rule-executing-thread"
	RemoteRulePoolSizeFactor  = 4
)

// LtThreadPoolFactory ports org.languagetool.tools.LtThreadPoolFactory
// as a named registry of LtThreadPoolExecutor instances.
type LtThreadPoolFactory struct {
	mu   sync.Mutex
	pools map[string]*LtThreadPoolExecutor
}

// DefaultLtThreadPoolFactory is the process-wide factory.
var DefaultLtThreadPoolFactory = &LtThreadPoolFactory{pools: map[string]*LtThreadPoolExecutor{}}

// CreateFixedThreadPoolExecutor creates or reuses a named pool.
func CreateFixedThreadPoolExecutor(identifier string, maxThreads, maxTaskInQueue int, reuse bool) *LtThreadPoolExecutor {
	return DefaultLtThreadPoolFactory.Create(identifier, maxThreads/2, maxThreads, maxTaskInQueue, reuse)
}

// Create creates a pool; core defaults to max/2 when core<=0.
func (f *LtThreadPoolFactory) Create(identifier string, core, maxThreads, maxTaskInQueue int, reuse bool) *LtThreadPoolExecutor {
	if f == nil {
		f = DefaultLtThreadPoolFactory
	}
	if core <= 0 {
		core = maxThreads / 2
		if core < 1 {
			core = 1
		}
	}
	if maxThreads < core {
		maxThreads = core
	}
	if maxTaskInQueue <= 0 {
		maxTaskInQueue = 64
	}
	if reuse {
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.pools == nil {
			f.pools = map[string]*LtThreadPoolExecutor{}
		}
		if p, ok := f.pools[identifier]; ok {
			return p
		}
		p := NewLtThreadPoolExecutor(identifier, maxThreads, maxTaskInQueue)
		p.Start()
		f.pools[identifier] = p
		return p
	}
	p := NewLtThreadPoolExecutor(identifier, maxThreads, maxTaskInQueue)
	p.Start()
	return p
}

// Get returns a previously created pool (nil if missing).
func (f *LtThreadPoolFactory) Get(identifier string) *LtThreadPoolExecutor {
	if f == nil {
		return nil
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.pools[identifier]
}

// ShutdownAll stops all registered pools.
func (f *LtThreadPoolFactory) ShutdownAll() {
	if f == nil {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, p := range f.pools {
		// best-effort; Shutdown waits
		go func(pool *LtThreadPoolExecutor) {
			// give in-flight work a moment
			time.Sleep(10 * time.Millisecond)
			pool.Shutdown()
		}(p)
	}
	// wait a bit then clear
	time.Sleep(50 * time.Millisecond)
	f.pools = map[string]*LtThreadPoolExecutor{}
}

// ResetLtThreadPoolFactory clears the default registry (tests).
func ResetLtThreadPoolFactory() {
	DefaultLtThreadPoolFactory = &LtThreadPoolFactory{pools: map[string]*LtThreadPoolExecutor{}}
}
