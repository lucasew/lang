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
	c.mu.Lock()
	defer c.mu.Unlock()
	if f, ok := c.cache[className]; ok {
		return f
	}
	factory, ok := c.factories[className]
	if !ok {
		panic(fmt.Sprintf("Could not find filter class: '%s' - make sure to use a fully qualified class name like 'org.languagetool.rules.MyFilter'", className))
	}
	filter := factory()
	if filter == nil {
		panic(fmt.Sprintf("Filter class '%s' factory returned nil", className))
	}
	c.cache[className] = filter
	return filter
}
