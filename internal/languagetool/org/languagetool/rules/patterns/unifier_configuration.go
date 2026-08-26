package patterns

import "sync"

// UnifierConfiguration ports org.languagetool.rules.patterns.UnifierConfiguration.
type UnifierConfiguration struct {
	mu                  sync.Mutex
	equivalenceTypes    map[EquivalenceTypeLocator]*PatternToken
	equivalenceFeatures map[string][]string
}

func NewUnifierConfiguration() *UnifierConfiguration {
	return &UnifierConfiguration{
		equivalenceTypes:    map[EquivalenceTypeLocator]*PatternToken{},
		equivalenceFeatures: map[string][]string{},
	}
}

// SetEquivalence registers a PatternToken for feature/type (no-op if already set).
func (c *UnifierConfiguration) SetEquivalence(feature, typ string, elem *PatternToken) {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := NewEquivalenceTypeLocator(feature, typ)
	if _, ok := c.equivalenceTypes[key]; ok {
		return
	}
	c.equivalenceTypes[key] = elem
	c.equivalenceFeatures[feature] = append(c.equivalenceFeatures[feature], typ)
}

func (c *UnifierConfiguration) GetEquivalenceTypes() map[EquivalenceTypeLocator]*PatternToken {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[EquivalenceTypeLocator]*PatternToken, len(c.equivalenceTypes))
	for k, v := range c.equivalenceTypes {
		out[k] = v
	}
	return out
}

func (c *UnifierConfiguration) GetEquivalenceFeatures() map[string][]string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string][]string, len(c.equivalenceFeatures))
	for k, v := range c.equivalenceFeatures {
		out[k] = append([]string(nil), v...)
	}
	return out
}

// CreateUnifier ports UnifierConfiguration.createUnifier.
func (c *UnifierConfiguration) CreateUnifier() *Unifier {
	return NewUnifier(c.GetEquivalenceTypes(), c.GetEquivalenceFeatures())
}
