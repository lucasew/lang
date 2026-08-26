package filters

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	c.Register("org.languagetool.rules.ar.filters.ArabicDateCheckFilter", func() patterns.RuleFilter {
		return NewArabicDateCheckFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicDMYDateCheckFilter", func() patterns.RuleFilter {
		return NewArabicDMYDateCheckFilter()
	})
	// ArabicAdvancedSynthesizer: empty subclass; wires Arabic synthesizer via ensureSynth.
	c.Register("org.languagetool.rules.ar.filters.ArabicAdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return NewArabicAdvancedSynthesizerFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicNumberPhraseFilter", func() patterns.RuleFilter {
		return NewArabicNumberPhraseFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicMasdarToVerbFilter", func() patterns.RuleFilter {
		return NewArabicMasdarToVerbFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicAdjectiveToExclamationFilter", func() patterns.RuleFilter {
		return NewArabicAdjectiveToExclamationFilter()
	})
	c.Register("org.languagetool.rules.ar.filters.ArabicVerbToMafoulMutlaqFilter", func() patterns.RuleFilter {
		return NewArabicVerbToMafoulMutlaqFilter()
	})
}
