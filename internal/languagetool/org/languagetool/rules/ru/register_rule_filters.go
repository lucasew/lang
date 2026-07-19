package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	// Suppress-misspelled: FilterDictIsMisspelled when ru_RU.dict wired (Java default spelling rule).
	c.Register("org.languagetool.rules.ru.RussianSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return NewRussianSuppressMisspelledSuggestionsFilter()
	})
	// Partial POS: NoDisambiguation uses process-wide tagger hook; RussianPartial needs disambiguator too.
	c.Register("org.languagetool.rules.ru.NoDisambiguationRussianPartialPosTagFilter", func() patterns.RuleFilter {
		return NewNoDisambiguationRussianPartialPosTagFilter(nil)
	})
	c.Register("org.languagetool.rules.ru.RussianPartialPosTagFilter", func() patterns.RuleFilter {
		// Fail-closed until tagger+disambiguator are both wired (do not invent disambiguation).
		return NewRussianPartialPosTagFilter(nil)
	})
	c.Register("org.languagetool.rules.ru.AdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return NewAdvancedSynthesizerFilter()
	})
	c.Register("org.languagetool.rules.ru.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
	c.Register("org.languagetool.rules.ru.FutureDateFilter", func() patterns.RuleFilter {
		return NewFutureDateFilter()
	})
}



