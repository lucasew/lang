package rules

import "sync"

// RemoteRuleFallbackManager ports org.languagetool.RemoteRuleFallbackManager
// (Java enum INSTANCE). Lives under rules for RemoteRuleConfig access; Java is
// org.languagetool.RemoteRuleFallbackManager.
type RemoteRuleFallbackManager struct {
	mu              sync.Mutex
	initCalled      bool
	fallbackConfigs map[string]*RemoteRuleConfig // primary ruleId → fallback config
}

// RemoteRuleFallbackInstance is the process singleton (Java INSTANCE).
var RemoteRuleFallbackInstance = &RemoteRuleFallbackManager{
	fallbackConfigs: map[string]*RemoteRuleConfig{},
}

// InitFromConfigs ports init(File) after configs are loaded — once only.
func (m *RemoteRuleFallbackManager) InitFromConfigs(configs []*RemoteRuleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.initCalled {
		return
	}
	m.setup(configs)
	m.initCalled = true
}

// InitForTests ports init_for_tests_only — resets and loads configs.
func (m *RemoteRuleFallbackManager) InitForTests(configs []*RemoteRuleConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.fallbackConfigs = map[string]*RemoteRuleConfig{}
	m.initCalled = true
	m.setup(configs)
}

func (m *RemoteRuleFallbackManager) setup(configs []*RemoteRuleConfig) {
	// Build id → config; for uniqueness, collect all with same id
	byID := map[string][]*RemoteRuleConfig{}
	for _, c := range configs {
		if c != nil {
			byID[c.GetRuleID()] = append(byID[c.GetRuleID()], c)
		}
	}
	for _, c := range configs {
		if c == nil {
			continue
		}
		fbID := c.GetFallbackRuleId()
		if fbID == "" {
			continue
		}
		// Java: filter rc.getRuleId().equals(fallbackRuleId); require size == 1
		matches := byID[fbID]
		if len(matches) != 1 {
			// skip when not found or not unique
			continue
		}
		m.fallbackConfigs[c.GetRuleID()] = matches[0]
	}
}

// GetInhouseFallback ports getInhouseFallback — nil if not init, empty, missing, or 3rd party.
func (m *RemoteRuleFallbackManager) GetInhouseFallback(ruleID string) *RemoteRuleConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.initCalled {
		return nil
	}
	if len(m.fallbackConfigs) == 0 {
		return nil
	}
	fb := m.fallbackConfigs[ruleID]
	if fb == nil || fb.IsUsingThirdPartyAI() {
		return nil
	}
	return fb
}

// GetFallback returns the fallback config for ruleId, or nil (test helper).
func (m *RemoteRuleFallbackManager) GetFallback(ruleID string) *RemoteRuleConfig {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.fallbackConfigs[ruleID]
}

// HasFallback reports whether a fallback is configured (raw map, ignores 3rd-party filter).
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

// CircuitBreakerState is a minimal OPEN/CLOSED stand-in for resilience4j.
type CircuitBreakerState string

const (
	CircuitClosed CircuitBreakerState = "CLOSED"
	CircuitOpen   CircuitBreakerState = "OPEN"
)

// RemoteRuleAvailability is the minimal surface for isRuleOrFallbackAvailable.
type RemoteRuleAvailability interface {
	GetID() string
	CircuitBreakerState() CircuitBreakerState
	GetFallbackRuleId() string
}

// IsRuleOrFallbackAvailable ports isRuleOrFallbackAvailable(rule, remoteRules).
// Returns: same rule id if available; fallback id if open with available fallback; "" if none.
func (m *RemoteRuleFallbackManager) IsRuleOrFallbackAvailable(rule RemoteRuleAvailability, remoteRules map[string]RemoteRuleAvailability) string {
	return m.isRuleOrFallbackAvailable(rule, remoteRules, map[string]struct{}{})
}

func (m *RemoteRuleFallbackManager) isRuleOrFallbackAvailable(rule RemoteRuleAvailability, remoteRules map[string]RemoteRuleAvailability, visited map[string]struct{}) string {
	if rule == nil {
		return ""
	}
	id := rule.GetID()
	if _, ok := visited[id]; ok {
		// circular fallback chain
		return ""
	}
	visited[id] = struct{}{}
	if rule.CircuitBreakerState() == CircuitOpen {
		fbID := rule.GetFallbackRuleId()
		if fbID != "" {
			if fb, ok := remoteRules[fbID]; ok && fb != nil {
				return m.isRuleOrFallbackAvailable(fb, remoteRules, visited)
			}
		}
		return ""
	}
	return id
}
