package rules

import "sync"

// RemoteRuleFallbackManager ports org.languagetool.RemoteRuleFallbackManager.
type RemoteRuleFallbackManager struct {
	mu              sync.Mutex
	initCalled      bool
	fallbackConfigs map[string]*RemoteRuleConfig // primary ruleId → fallback config
}

// RemoteRuleFallbackInstance is the process singleton (Java INSTANCE).
var RemoteRuleFallbackInstance = &RemoteRuleFallbackManager{
	fallbackConfigs: map[string]*RemoteRuleConfig{},
}

// InitFromConfigs builds fallback map (idempotent like Java init once).
func (m *RemoteRuleFallbackManager) InitFromConfigs(configs []*RemoteRuleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.initCalled {
		return
	}
	m.setup(configs)
	m.initCalled = true
}

// InitForTests resets and loads configs (Java init_for_tests_only).
func (m *RemoteRuleFallbackManager) InitForTests(configs []*RemoteRuleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallbackConfigs = map[string]*RemoteRuleConfig{}
	m.initCalled = true
	m.setup(configs)
}

func (m *RemoteRuleFallbackManager) setup(configs []*RemoteRuleConfig) {
	byID := map[string]*RemoteRuleConfig{}
	for _, c := range configs {
		if c != nil {
			byID[c.GetRuleID()] = c
		}
	}
	for _, c := range configs {
		if c == nil {
			continue
		}
		fbID := ""
		if c.Options != nil {
			fbID = c.Options[RemoteOptionFallbackRule]
		}
		// also support field if present on config via options only (Java getFallbackRuleId)
		if fbID == "" {
			continue
		}
		fb, ok := byID[fbID]
		if !ok {
			continue
		}
		m.fallbackConfigs[c.GetRuleID()] = fb
	}
}

// GetFallback returns the fallback config for ruleId, or nil.
func (m *RemoteRuleFallbackManager) GetFallback(ruleID string) *RemoteRuleConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.fallbackConfigs[ruleID]
}

// HasFallback reports whether a fallback is configured.
func (m *RemoteRuleFallbackManager) HasFallback(ruleID string) bool {
	return m.GetFallback(ruleID) != nil
}

// Clear resets state (tests).
func (m *RemoteRuleFallbackManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallbackConfigs = map[string]*RemoteRuleConfig{}
	m.initCalled = false
}
