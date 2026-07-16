package server

import (
	"sync"
)

// PipelinePool ports org.languagetool.server.PipelinePool as a simple keyed pool.
// When caching is disabled, Create always builds a fresh pipeline.
type PipelinePool struct {
	mu     sync.Mutex
	cfg    *HTTPServerConfig
	// key → idle pipelines
	idle map[string][]*Pipeline
	// outstanding borrowed count (optional metrics)
	borrowed int
}

func NewPipelinePool(cfg *HTTPServerConfig) *PipelinePool {
	if cfg == nil {
		cfg = NewHTTPServerConfig()
	}
	return &PipelinePool{
		cfg:  cfg,
		idle: map[string][]*Pipeline{},
	}
}

// Create builds a configured pipeline for settings (always new; used by factory).
func (p *PipelinePool) Create(settings PipelineSettings) *Pipeline {
	pl := NewPipeline(settings)
	// Apply config-level disabled rules before freeze.
	if p != nil && p.cfg != nil {
		for _, id := range p.cfg.DisabledRuleIDs {
			_ = pl.DisableRule(id)
		}
		if p.cfg.MaxErrorsPerWordRate > 0 {
			_ = pl.SetMaxErrorsPerWordRate(p.cfg.MaxErrorsPerWordRate)
		}
	}
	if settings.Query.UseQuerySettings {
		for _, id := range settings.Query.DisabledRules {
			_ = pl.DisableRule(id)
		}
		for _, id := range settings.Query.EnabledRules {
			_ = pl.EnableRule(id)
		}
	}
	if p != nil && p.cfg != nil && p.cfg.IsPipelineCachingEnabled() {
		pl.SetupFinished()
	}
	return pl
}

// Borrow returns a pipeline for settings (from pool or newly created).
func (p *PipelinePool) Borrow(settings PipelineSettings) (*Pipeline, error) {
	if p == nil {
		return NewPipeline(settings), nil
	}
	if !p.cfg.IsPipelineCachingEnabled() {
		return p.Create(settings), nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	key := settings.Key()
	if list := p.idle[key]; len(list) > 0 {
		pl := list[len(list)-1]
		p.idle[key] = list[:len(list)-1]
		p.borrowed++
		return pl, nil
	}
	// capacity check
	max := p.cfg.GetMaxPipelinePoolSize()
	totalIdle := 0
	for _, list := range p.idle {
		totalIdle += len(list)
	}
	if p.borrowed+totalIdle >= max {
		return nil, NewUnavailableError("pipeline pool exhausted", nil)
	}
	pl := p.Create(settings)
	p.borrowed++
	return pl, nil
}

// Return puts a pipeline back into the pool (no-op if caching disabled).
func (p *PipelinePool) Return(settings PipelineSettings, pl *Pipeline) {
	if p == nil || pl == nil || !p.cfg.IsPipelineCachingEnabled() {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	key := settings.Key()
	p.idle[key] = append(p.idle[key], pl)
	if p.borrowed > 0 {
		p.borrowed--
	}
}

// IdleCount returns number of idle pipelines for a settings key.
func (p *PipelinePool) IdleCount(settings PipelineSettings) int {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.idle[settings.Key()])
}

// Borrowed returns currently borrowed pipeline count.
func (p *PipelinePool) Borrowed() int {
	if p == nil {
		return 0
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.borrowed
}
