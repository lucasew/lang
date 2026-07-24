package tools

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// LtThreadPoolExecutor ports org.languagetool.tools.LtThreadPoolExecutor
// as a bounded worker pool with queue metrics (no Prometheus dependency).
type LtThreadPoolExecutor struct {
	name        string
	workers     int
	queue       chan func()
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	started     atomic.Bool
	queueLen    atomic.Int64
	maxQueue    int
	largestPool atomic.Int64
	active      atomic.Int64
	exitOnOOM   bool
}

// NewLtThreadPoolExecutor creates a pool; call Start to run workers.
func NewLtThreadPoolExecutor(name string, coreWorkers, maxQueue int) *LtThreadPoolExecutor {
	if coreWorkers <= 0 {
		coreWorkers = 1
	}
	if maxQueue <= 0 {
		maxQueue = 64
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &LtThreadPoolExecutor{
		name:     name,
		workers:  coreWorkers,
		queue:    make(chan func(), maxQueue),
		ctx:      ctx,
		cancel:   cancel,
		maxQueue: maxQueue,
	}
}

func (p *LtThreadPoolExecutor) Name() string { return p.name }

func (p *LtThreadPoolExecutor) MaxQueueSize() int { return p.maxQueue }

func (p *LtThreadPoolExecutor) QueueSize() int {
	return int(p.queueLen.Load())
}

func (p *LtThreadPoolExecutor) LargestPoolSize() int64 {
	return p.largestPool.Load()
}

// Start launches worker goroutines.
func (p *LtThreadPoolExecutor) Start() {
	if p == nil || !p.started.CompareAndSwap(false, true) {
		return
	}
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.loop()
	}
}

func (p *LtThreadPoolExecutor) loop() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case job, ok := <-p.queue:
			if !ok {
				return
			}
			p.queueLen.Add(-1)
			p.active.Add(1)
			if cur := p.active.Load(); cur > p.largestPool.Load() {
				p.largestPool.Store(cur)
			}
			func() {
				defer p.active.Add(-1)
				if job != nil {
					job()
				}
			}()
		}
	}
}

// Execute enqueues a job; returns false if the queue is full or pool is stopped.
func (p *LtThreadPoolExecutor) Execute(job func()) bool {
	if p == nil || job == nil {
		return false
	}
	if !p.started.Load() {
		p.Start()
	}
	select {
	case <-p.ctx.Done():
		return false
	case p.queue <- job:
		p.queueLen.Add(1)
		return true
	default:
		return false
	}
}

// ExecuteWait enqueues with a timeout.
func (p *LtThreadPoolExecutor) ExecuteWait(job func(), timeout time.Duration) bool {
	if p == nil || job == nil {
		return false
	}
	if !p.started.Load() {
		p.Start()
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-p.ctx.Done():
		return false
	case p.queue <- job:
		p.queueLen.Add(1)
		return true
	case <-timer.C:
		return false
	}
}

// Shutdown stops accepting work and waits for workers.
func (p *LtThreadPoolExecutor) Shutdown() {
	if p == nil {
		return
	}
	p.cancel()
	// drain is optional; workers exit on ctx
	p.wg.Wait()
}
