package patterns

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RegexRuleFilterCreator ports org.languagetool.rules.patterns.RegexRuleFilterCreator
// as a name→factory registry (no Java class loading).
type RegexRuleFilterCreator struct {
	factories map[string]func() RegexRuleFilter
}

func NewRegexRuleFilterCreator() *RegexRuleFilterCreator {
	c := &RegexRuleFilterCreator{factories: map[string]func() RegexRuleFilter{}}
	c.Register("org.languagetool.rules.patterns.RegexAntiPatternFilter", func() RegexRuleFilter {
		return regexAntiPatternAsRegexFilter{}
	})
	return c
}

// Register adds a factory under fully-qualified class name.
func (c *RegexRuleFilterCreator) Register(className string, factory func() RegexRuleFilter) {
	c.factories[className] = factory
}

// GetFilter returns a new filter instance for className.
func (c *RegexRuleFilterCreator) GetFilter(className string) RegexRuleFilter {
	f, ok := c.factories[className]
	if !ok {
		panic(fmt.Sprintf("Could not find filter class: '%s' - register it on RegexRuleFilterCreator", className))
	}
	return f()
}

// Adapt RegexAntiPatternFilter to RegexRuleFilter (patternMatch unused).
type regexAntiPatternAsRegexFilter struct{}

func (regexAntiPatternAsRegexFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string,
	sentence *languagetool.AnalyzedSentence, _ string) *rules.RuleMatch {
	return (RegexAntiPatternFilter{}).AcceptRegexMatch(match, arguments, sentence)
}
