package en

import (
	"fmt"
	"strings"
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
	c.Register("org.languagetool.rules.en.NewYearDateFilter", func() patterns.RuleFilter {
		return newYearDateRuleFilter{core: enNewYearDateCore()}
	})
	c.Register("org.languagetool.rules.en.YMDNewYearDateFilter", func() patterns.RuleFilter {
		return ymdNewYearDateRuleFilter{core: enNewYearDateCore(), ymd: rules.NewYMDDateHelper()}
	})
	// Suppress-misspelled: without a speller hook, nothing is treated as misspelled
	// (Java AbstractSuppressMisspelledSuggestionsFilter with no spelling rule keeps suggestions).
	c.Register("org.languagetool.rules.en.EnglishSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return suppressMisspelledRuleFilter{inner: &rules.AbstractSuppressMisspelledSuggestionsFilter{}}
	})
	// Number-in-word / FindSuggestions need a Morfologik speller. Fail-closed until wired:
	// empty replacements / nil speller → drop match (do not invent suggestions).
	c.Register("org.languagetool.rules.en.EnglishNumberInWordFilter", func() patterns.RuleFilter {
		return numberInWordRuleFilter{}
	})
	c.Register("org.languagetool.rules.en.FindSuggestionsFilter", func() patterns.RuleFilter {
		return findSuggestionsRuleFilter{inner: &rules.AbstractFindSuggestionsFilter{}}
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

func enNewYearDateCore() *rules.NewYearDateFilterCore {
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

// newYearDateRuleFilter ports AbstractNewYearDateFilter.
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

// ymdNewYearDateRuleFilter ports YMDNewYearDateFilter (date=yyyy-mm-dd then NewYear).
type ymdNewYearDateRuleFilter struct {
	core *rules.NewYearDateFilterCore
	ymd  *rules.YMDDateHelper
}

func (f ymdNewYearDateRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil || f.core == nil || f.ymd == nil {
		return nil
	}
	if _, ok := arguments["year"]; ok {
		panic("set only 'weekDay' and 'date' for YMDNewYearDateFilter")
	}
	if _, ok := arguments["month"]; ok {
		panic("set only 'weekDay' and 'date' for YMDNewYearDateFilter")
	}
	if _, ok := arguments["day"]; ok {
		panic("set only 'weekDay' and 'date' for YMDNewYearDateFilter")
	}
	parsed, err := f.ymd.ParseDate(arguments)
	if err != nil {
		return nil
	}
	// Java: correctDate replaces {realDate} with year+1-mm-dd before NewYear filter.
	y := 0
	_, _ = fmt.Sscanf(parsed["year"], "%d", &y)
	correctDate := fmt.Sprintf("%d-%s-%s", y+1, parsed["month"], parsed["day"])
	msg := match.GetMessage()
	msg = strings.ReplaceAll(msg, "{realDate}", correctDate)
	m2 := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	m2.ShortMessage = match.ShortMessage
	return (newYearDateRuleFilter{core: f.core}).AcceptRuleMatch(m2, parsed, 0, nil, nil)
}

// suppressMisspelledRuleFilter ports AbstractSuppressMisspelledSuggestionsFilter.
type suppressMisspelledRuleFilter struct {
	inner *rules.AbstractSuppressMisspelledSuggestionsFilter
}

func (f suppressMisspelledRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments)
}

// numberInWordRuleFilter ports AbstractNumberInWordFilter without a speller:
// only pure surface rewrites are considered if they differ; without isMisspelled
// we cannot accept invented dictionary hits, so drop (fail-closed).
type numberInWordRuleFilter struct{}

func (numberInWordRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	// Java requires isMisspelled + getSuggestions from MorfologikAmericanSpellerRule.
	// Incomplete stack: do not invent suggestions; drop match.
	_ = match
	_ = arguments
	return nil
}

// findSuggestionsRuleFilter ports FindSuggestionsFilter; nil SpellingSuggestions → drop.
type findSuggestionsRuleFilter struct {
	inner *rules.AbstractFindSuggestionsFilter
}

func (f findSuggestionsRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
