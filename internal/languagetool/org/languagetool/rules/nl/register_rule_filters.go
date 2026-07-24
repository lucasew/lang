package nl

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	// Suppress-misspelled: FilterDictIsMisspelled when nl_NL.dict wired (Java default spelling rule).
	c.Register("org.languagetool.rules.nl.DutchSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return NewDutchSuppressMisspelledSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.nl.CompoundFilter", func() patterns.RuleFilter {
		return compoundRuleFilter{}
	})
	c.Register("org.languagetool.rules.nl.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
	c.Register("org.languagetool.rules.nl.NewYearDateFilter", func() patterns.RuleFilter {
		return newYearDateRuleFilter{core: nlNewYearDateCore()}
	})
	// Number-in-word: MorfologikDutchSpellerRule via WireDutchFilterSpeller; fail-closed without dict.
	c.Register("org.languagetool.rules.nl.DutchNumberInWordFilter", func() patterns.RuleFilter {
		return NewDutchNumberInWordFilter()
	})
}

type compoundRuleFilter struct{}

func (compoundRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	return NewCompoundFilter().AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}

func nlDateCheckWithSuggestions() *rules.AbstractDateCheckWithSuggestionsFilter {
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
			FormatDayOfWeek: formatNLDayOfWeek,
			GetMonth: func(localized string) int {
				m, err := h.GetMonth(localized)
				if err != nil {
					return 0
				}
				return int(m)
			},
		},
		// Java DateCheckFilter.getErrorMessageWrongYear
		ErrorMessageWrongYear: `Deze datum is onjuist. Bedoelt u misschien "{currentYear}"?`,
	}
}

func nlNewYearDateCore() *rules.NewYearDateFilterCore {
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

type newYearDateRuleFilter struct {
	core *rules.NewYearDateFilterCore
}

func (f newYearDateRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil || f.core == nil {
		return nil
	}
	msg := f.core.AcceptFromArgs(arguments, match.GetMessage())
	if msg == "" {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	out.ShortMessage = match.ShortMessage
	return out
}
