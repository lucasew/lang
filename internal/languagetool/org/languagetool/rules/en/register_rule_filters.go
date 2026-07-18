package en

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register Java EN RuleFilter class names so grammar.xml can load those rules.
// Unknown filters remain skipped (fail-closed).
func init() {
	c := patterns.GlobalRuleFilterCreator
	c.Register("org.languagetool.rules.en.OrdinalSuffixFilter", func() patterns.RuleFilter {
		return ordinalSuffixRuleFilter{}
	})
	c.Register("org.languagetool.rules.en.AdverbFilter", func() patterns.RuleFilter {
		return adverbRuleFilter{}
	})
	c.Register("org.languagetool.rules.en.FutureDateFilter", func() patterns.RuleFilter {
		return futureDateRuleFilter{core: enFutureDateCore()}
	})
	c.Register("org.languagetool.rules.en.DateCheckFilter", func() patterns.RuleFilter {
		return dateCheckRuleFilter{inner: enDateCheckWithSuggestions()}
	})
}

// ordinalSuffixRuleFilter ports OrdinalSuffixFilter.acceptRuleMatch.
type ordinalSuffixRuleFilter struct{}

func (ordinalSuffixRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, _ map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	reps := match.GetSuggestedReplacements()
	if len(reps) == 0 {
		// Java get(0) would throw; without a suggestion there is nothing to fix.
		return match
	}
	fixed := NewOrdinalSuffixFilter().Fix(reps[0])
	match.SetSuggestedReplacement(fixed)
	return match
}

// adverbRuleFilter ports AdverbFilter.acceptRuleMatch.
type adverbRuleFilter struct{}

func (adverbRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	adverb := arguments["adverb"]
	noun := arguments["noun"]
	if sug := NewAdverbFilter().Suggest(adverb, noun); sug != "" {
		match.SetSuggestedReplacement(sug)
	}
	return match
}

func enFutureDateCore() *rules.FutureDateFilterCore {
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

// futureDateRuleFilter ports AbstractFutureDateFilter: keep match only if date is in the future.
type futureDateRuleFilter struct {
	core *rules.FutureDateFilterCore
}

func (f futureDateRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil || f.core == nil {
		return nil
	}
	if f.core.AcceptFromArgs(arguments) {
		return match
	}
	return nil
}

func enDateCheckWithSuggestions() *rules.AbstractDateCheckWithSuggestionsFilter {
	h := NewDateFilterHelper()
	return &rules.AbstractDateCheckWithSuggestionsFilter{
		AbstractDateCheckFilter: rules.AbstractDateCheckFilter{
			GetDayOfWeekName: func(localized string) time.Weekday {
				wd, err := h.GetDayOfWeek(localized)
				if err != nil {
					// Java DateFilterHelper throws RuntimeException on unknown weekday.
					panic(err)
				}
				return wd
			},
			FormatDayOfWeek: func(t time.Time) string {
				// Java DateFilterHelper Locale.UK LONG: Sunday, Monday, …
				names := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
				return names[int(t.Weekday())]
			},
			GetMonth: func(localized string) int {
				m, err := h.GetMonth(localized)
				if err != nil {
					return 0
				}
				return int(m)
			},
		},
		ErrorMessageWrongYear: `This date is wrong. Did you mean "{currentYear}"?`,
	}
}

// dateCheckRuleFilter ports DateCheckFilter (AbstractDateCheckWithSuggestionsFilter).
type dateCheckRuleFilter struct {
	inner *rules.AbstractDateCheckWithSuggestionsFilter
}

func (f dateCheckRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
