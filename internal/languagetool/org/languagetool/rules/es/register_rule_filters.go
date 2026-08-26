package es

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	// Suppress-misspelled: FilterDictIsMisspelled when es-ES.dict wired (Java default spelling rule).
	c.Register("org.languagetool.rules.es.SpanishSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return NewSpanishSuppressMisspelledSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.es.ConfusionCheckFilter", func() patterns.RuleFilter {
		return confusionCheckRuleFilter{inner: NewConfusionCheckFilter().ConfusionCheckFilter}
	})
	c.Register("org.languagetool.rules.es.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
	c.Register("org.languagetool.rules.es.NewYearDateFilter", func() patterns.RuleFilter {
		return NewNewYearDateFilter()
	})
	// FindSuggestionsFilter: Morfologik findSimilarWords via WireSpanishFilterSpeller; Tag optional.
	c.Register("org.languagetool.rules.es.FindSuggestionsFilter", func() patterns.RuleFilter {
		return NewFindSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.es.AdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return NewAdvancedSynthesizerFilter()
	})
	// Number-in-word: MorfologikSpanishSpellerRule via WireSpanishFilterSpeller; fail-closed without dict.
	c.Register("org.languagetool.rules.es.SpanishNumberInWordFilter", func() patterns.RuleFilter {
		return NewSpanishNumberInWordFilter()
	})
	// Full AcceptRuleMatch port; wire SpanishSynthesizer via Synthesize when available.
	c.Register("org.languagetool.rules.es.PostponedAdjectiveConcordanceFilter", func() patterns.RuleFilter {
		return NewPostponedAdjectiveConcordanceFilter()
	})
}



type confusionCheckRuleFilter struct {
	inner *rules.ConfusionCheckFilter
}

func (f confusionCheckRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

func esDateCheckWithSuggestions() *rules.AbstractDateCheckWithSuggestionsFilter {
	h := NewDateFilterHelper()
	return &rules.AbstractDateCheckWithSuggestionsFilter{
		AbstractDateCheckFilter: rules.AbstractDateCheckFilter{
			GetDayOfWeekName: func(localized string) time.Weekday {
				wd, err := h.GetDayOfWeek(localized)
				if err != nil {
					panic(err)
				}
				return wd
			},
			FormatDayOfWeek: formatESDayOfWeek,
			GetMonth: func(localized string) int {
				m, err := h.GetMonth(localized)
				if err != nil {
					return 0
				}
				return int(m)
			},
		},
		// Java DateCheckFilter.getErrorMessageWrongYear (upstream typo "Este fecha" preserved)
		ErrorMessageWrongYear: `Este fecha no es correcta. ¿Se refería al año "{currentYear}"?`,
	}
}

func esNewYearDateCore() *rules.NewYearDateFilterCore {
	h := NewDateFilterHelper()
	return &rules.NewYearDateFilterCore{
		GetMonth: func(localized string) (int, error) {
			m, err := h.GetMonth(localized)
			if err != nil {
				return 0, err
			}
			return int(m), nil
		},
	}
}

