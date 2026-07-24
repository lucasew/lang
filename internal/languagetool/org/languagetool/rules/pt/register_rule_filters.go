package pt

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	// Regex-rule filters (regexp XML rules).
	patterns.GlobalRegexRuleFilterCreator.Register(
		"org.languagetool.rules.pt.BrazilianToponymFilter",
		func() patterns.RegexRuleFilter { return NewBrazilianToponymFilter() },
	)
	c := patterns.GlobalRuleFilterCreator
	// Suppress-misspelled: FilterDictIsMisspelled when pt spelling dict wired (Java default spelling rule).
	c.Register("org.languagetool.rules.pt.PortugueseSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return NewPortugueseSuppressMisspelledSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.pt.ConfusionCheckFilter", func() patterns.RuleFilter {
		return confusionCheckRuleFilter{inner: NewConfusionCheckFilter().ConfusionCheckFilter}
	})
	c.Register("org.languagetool.rules.pt.FutureDateFilter", func() patterns.RuleFilter {
		return NewFutureDateFilter()
	})
	c.Register("org.languagetool.rules.pt.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
	c.Register("org.languagetool.rules.pt.YMDDateCheckFilter", func() patterns.RuleFilter {
		return NewYMDDateCheckFilter()
	})
	c.Register("org.languagetool.rules.pt.NewYearDateFilter", func() patterns.RuleFilter {
		return NewNewYearDateFilter()
	})
	c.Register("org.languagetool.rules.pt.YMDNewYearDateFilter", func() patterns.RuleFilter {
		return NewYMDNewYearDateFilter()
	})
	c.Register("org.languagetool.rules.pt.AdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return NewAdvancedSynthesizerFilter()
	})
	c.Register("org.languagetool.rules.pt.RomanNumeralFilter", func() patterns.RuleFilter {
		return NewRomanNumeralFilter()
	})
	// Enclisis/proclisis need Portuguese synthesizer for full forms; register fail-closed without.
	c.Register("org.languagetool.rules.pt.PortugueseEnclisisFilter", func() patterns.RuleFilter {
		return NewPortugueseEnclisisFilter()
	})
	c.Register("org.languagetool.rules.pt.PortugueseProclisisFilter", func() patterns.RuleFilter {
		return NewPortugueseProclisisFilter()
	})
	// Participle regular/irregular needs PT synthesizer; fail-closed without Synthesize.
	c.Register("org.languagetool.rules.pt.RegularIrregularParticipleFilter", func() patterns.RuleFilter {
		return NewRegularIrregularParticipleFilter()
	})
	// Partial POS needs PT tagger; Tag nil → fail-closed.
	c.Register("org.languagetool.rules.pt.NoDisambiguationPortuguesePartialPosTagFilter", func() patterns.RuleFilter {
		return NewNoDisambiguationPortuguesePartialPosTagFilter(nil)
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

func ptFutureDateCore() *rules.FutureDateFilterCore {
	h := NewDateFilterHelper()
	return &rules.FutureDateFilterCore{
		GetMonth: func(localized string) (int, error) {
			m, err := h.GetMonth(localized)
			if err != nil {
				return 0, err
			}
			return int(m), nil
		},
	}
}


func ptDateCheckWithSuggestions() *rules.AbstractDateCheckWithSuggestionsFilter {
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
			FormatDayOfWeek: formatPTDayOfWeek,
			GetMonth: func(localized string) int {
				m, err := h.GetMonth(localized)
				if err != nil {
					return 0
				}
				return int(m)
			},
		},
		// Java DateCheckFilter.getErrorMessageWrongYear
		ErrorMessageWrongYear: `Esta data está incorreta. Você está se referindo ao ano "{currentYear}"?`,
	}
}

func ptNewYearDateCore() *rules.NewYearDateFilterCore {
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

