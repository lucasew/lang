package patterns

import (
	"fmt"
	"sync"
)

// RuleFilterCreator ports org.languagetool.rules.patterns.RuleFilterCreator
// as a name→factory registry (no Class.forName).
type RuleFilterCreator struct {
	mu        sync.Mutex
	factories map[string]func() RuleFilter
	cache     map[string]RuleFilter
}

// GlobalRuleFilterCreator is the process singleton (ports getInstance).
var GlobalRuleFilterCreator = NewRuleFilterCreator()

func NewRuleFilterCreator() *RuleFilterCreator {
	return &RuleFilterCreator{
		factories: map[string]func() RuleFilter{},
		cache:     map[string]RuleFilter{},
	}
}

// Register associates a fully-qualified class name with a filter factory.
func (c *RuleFilterCreator) Register(className string, factory func() RuleFilter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[className] = factory
}

// GetFilter ports RuleFilterCreator.getFilter.
func (c *RuleFilterCreator) GetFilter(className string) RuleFilter {
	f, ok := c.TryGetFilter(className)
	if !ok {
		panic(fmt.Sprintf("Could not find filter class: '%s' - make sure to use a fully qualified class name like 'org.languagetool.rules.MyFilter'", className))
	}
	return f
}

// TryGetFilter returns a filter when the class is registered; false if unknown or factory is nil.
// Used by grammar load so missing filters skip the rule (fail-closed) instead of panicking.
func (c *RuleFilterCreator) TryGetFilter(className string) (RuleFilter, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if f, ok := c.cache[className]; ok {
		return f, true
	}
	factory, ok := c.factories[className]
	if !ok {
		return nil, false
	}
	filter := factory()
	if filter == nil {
		return nil, false
	}
	c.cache[className] = filter
	return filter, true
}

// HasFilter reports whether className is registered (without instantiating).
func (c *RuleFilterCreator) HasFilter(className string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.factories[className]
	return ok
}
