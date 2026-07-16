package server

// Pipeline ports org.languagetool.server.Pipeline as a freezeable check engine holder.
// Full JLanguageTool inheritance is deferred; this twin holds rule enable/disable state
// and language identity for pool reuse.
type Pipeline struct {
	settings PipelineSettings
	frozen   bool

	disabledRules map[string]struct{}
	enabledRules  map[string]struct{}
	cleanOverlaps bool
	maxErrRate    float64
}

func NewPipeline(settings PipelineSettings) *Pipeline {
	return &Pipeline{
		settings:      settings,
		disabledRules: map[string]struct{}{},
		enabledRules:  map[string]struct{}{},
		cleanOverlaps: true,
	}
}

// SetupFinished freezes the pipeline against further mutation (pool safety).
func (p *Pipeline) SetupFinished() {
	if p != nil {
		p.frozen = true
	}
}

func (p *Pipeline) IsFrozen() bool { return p != nil && p.frozen }

func (p *Pipeline) Settings() PipelineSettings {
	if p == nil {
		return PipelineSettings{}
	}
	return p.settings
}

func (p *Pipeline) preventModification() error {
	if p != nil && p.frozen {
		return &IllegalPipelineMutationError{}
	}
	return nil
}

func (p *Pipeline) SetCleanOverlappingMatches(v bool) error {
	if err := p.preventModification(); err != nil {
		return err
	}
	p.cleanOverlaps = v
	return nil
}

func (p *Pipeline) SetMaxErrorsPerWordRate(v float64) error {
	if err := p.preventModification(); err != nil {
		return err
	}
	p.maxErrRate = v
	return nil
}

func (p *Pipeline) DisableRule(ruleID string) error {
	if err := p.preventModification(); err != nil {
		return err
	}
	p.disabledRules[ruleID] = struct{}{}
	delete(p.enabledRules, ruleID)
	return nil
}

func (p *Pipeline) DisableRules(ruleIDs []string) error {
	for _, id := range ruleIDs {
		if err := p.DisableRule(id); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) EnableRule(ruleID string) error {
	if err := p.preventModification(); err != nil {
		return err
	}
	p.enabledRules[ruleID] = struct{}{}
	delete(p.disabledRules, ruleID)
	return nil
}

func (p *Pipeline) DisabledRules() []string {
	if p == nil {
		return nil
	}
	out := make([]string, 0, len(p.disabledRules))
	for id := range p.disabledRules {
		out = append(out, id)
	}
	return out
}

func (p *Pipeline) IsRuleDisabled(ruleID string) bool {
	if p == nil {
		return false
	}
	_, ok := p.disabledRules[ruleID]
	return ok
}

func (p *Pipeline) CleanOverlappingMatches() bool {
	return p == nil || p.cleanOverlaps
}

func (p *Pipeline) MaxErrorsPerWordRate() float64 {
	if p == nil {
		return 0
	}
	return p.maxErrRate
}
