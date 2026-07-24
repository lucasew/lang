package patterns

import (
	"fmt"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegexRuleFilterCreator ports org.languagetool.rules.patterns.RegexRuleFilterCreator
// as a name→factory registry (no Java class loading).
type RegexRuleFilterCreator struct {
	mu        sync.Mutex
	factories map[string]func() RegexRuleFilter
}

// GlobalRegexRuleFilterCreator is the process singleton for regexp-rule filters.
var GlobalRegexRuleFilterCreator = NewRegexRuleFilterCreator()

func NewRegexRuleFilterCreator() *RegexRuleFilterCreator {
	c := &RegexRuleFilterCreator{factories: map[string]func() RegexRuleFilter{}}
	c.Register("org.languagetool.rules.patterns.RegexAntiPatternFilter", func() RegexRuleFilter {
		return regexAntiPatternAsRegexFilter{}
	})
	return c
}

// Register adds a factory under fully-qualified class name.
func (c *RegexRuleFilterCreator) Register(className string, factory func() RegexRuleFilter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.factories[className] = factory
}

// GetFilter returns a new filter instance for className.
func (c *RegexRuleFilterCreator) GetFilter(className string) RegexRuleFilter {
	c.mu.Lock()
	defer c.mu.Unlock()
	f, ok := c.factories[className]
	if !ok {
		panic(fmt.Sprintf("Could not find filter class: '%s' - register it on RegexRuleFilterCreator", className))
	}
	return f()
}

// HasFilter reports whether className is registered.
func (c *RegexRuleFilterCreator) HasFilter(className string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.factories[className]
	return ok
}

// TryGetFilter returns a filter when registered (fail-closed for grammar load).
func (c *RegexRuleFilterCreator) TryGetFilter(className string) (RegexRuleFilter, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	factory, ok := c.factories[className]
	if !ok {
		return nil, false
	}
	return factory(), true
}

// Adapt RegexAntiPatternFilter to RegexRuleFilter (groups unused).
type regexAntiPatternAsRegexFilter struct{}

func (regexAntiPatternAsRegexFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string,
	sentence *languagetool.AnalyzedSentence, _ []string) *rules.RuleMatch {
	return (RegexAntiPatternFilter{}).AcceptRegexMatch(match, arguments, sentence)
}
