package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	// Partial POS needs Irish tagger; Tag nil → fail-closed (drop match).
	c.Register("org.languagetool.rules.ga.IrishPartialPosTagFilter", func() patterns.RuleFilter {
		return NewIrishPartialPosTagFilter(nil)
	})
	c.Register("org.languagetool.rules.ga.NoDisambiguationIrishPartialPosTagFilter", func() patterns.RuleFilter {
		return NewNoDisambiguationIrishPartialPosTagFilter(nil)
	})
}
