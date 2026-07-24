package server

import "sync"

// ActiveRules ports org.languagetool.server.ActiveRules / ActiveRulesMBean.
// Tracks in-flight pattern rules and spell checks for JMX-like introspection.
type ActiveRules struct {
	mu            sync.Mutex
	patternRules  map[string]int
	spellChecks   []string
	maxSpellQueue int
}

func NewActiveRules() *ActiveRules {
	return &ActiveRules{
		patternRules:  map[string]int{},
		spellChecks:   nil,
		maxSpellQueue: 1000,
	}
}

func (a *ActiveRules) GetActivePatternRules() map[string]int {
	if a == nil {
		return map[string]int{}
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make(map[string]int, len(a.patternRules))
	for k, v := range a.patternRules {
		out[k] = v
	}
	return out
}

func (a *ActiveRules) GetActiveSpellChecks() []string {
	if a == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]string, len(a.spellChecks))
	copy(out, a.spellChecks)
	return out
}

// EnterPattern increments the in-flight counter for a pattern rule id.
func (a *ActiveRules) EnterPattern(ruleID string) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.patternRules[ruleID]++
}

// LeavePattern decrements the in-flight counter for a pattern rule id.
func (a *ActiveRules) LeavePattern(ruleID string) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.patternRules[ruleID] <= 1 {
		delete(a.patternRules, ruleID)
	} else {
		a.patternRules[ruleID]--
	}
}

// EnterSpellCheck records an active spell check word.
func (a *ActiveRules) EnterSpellCheck(word string) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if len(a.spellChecks) >= a.maxSpellQueue {
		a.spellChecks = a.spellChecks[1:]
	}
	a.spellChecks = append(a.spellChecks, word)
}

// LeaveSpellCheck removes one occurrence of word from the active queue.
func (a *ActiveRules) LeaveSpellCheck(word string) {
	if a == nil {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i, w := range a.spellChecks {
		if w == word {
			a.spellChecks = append(a.spellChecks[:i], a.spellChecks[i+1:]...)
			return
		}
	}
}
