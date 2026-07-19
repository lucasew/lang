package en

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register Java EN RuleFilter class names so grammar.xml can load those rules.
// Unknown filters remain skipped (fail-closed).
func init() {
	c := patterns.GlobalRuleFilterCreator
	c.Register("org.languagetool.rules.en.OrdinalSuffixFilter", func() patterns.RuleFilter {
		return NewOrdinalSuffixFilter()
	})
	c.Register("org.languagetool.rules.en.AdverbFilter", func() patterns.RuleFilter {
		return NewAdverbFilter()
	})
	c.Register("org.languagetool.rules.en.FutureDateFilter", func() patterns.RuleFilter {
		return NewFutureDateFilter()
	})
	c.Register("org.languagetool.rules.en.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
	c.Register("org.languagetool.rules.en.YMDDateCheckFilter", func() patterns.RuleFilter {
		return NewYMDDateCheckFilter()
	})
	c.Register("org.languagetool.rules.en.NewYearDateFilter", func() patterns.RuleFilter {
		return NewNewYearDateFilter()
	})
	c.Register("org.languagetool.rules.en.YMDNewYearDateFilter", func() patterns.RuleFilter {
		return NewYMDNewYearDateFilter()
	})
	// Suppress-misspelled: FilterDictIsMisspelled when en_US.dict wired (Java default spelling rule).
	c.Register("org.languagetool.rules.en.EnglishSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return NewEnglishSuppressMisspelledSuggestionsFilter()
	})
	// Number-in-word / FindSuggestions: fail-closed without dict; full Java logic when wired.
	c.Register("org.languagetool.rules.en.EnglishNumberInWordFilter", func() patterns.RuleFilter {
		return NewEnglishNumberInWordFilter()
	})
	c.Register("org.languagetool.rules.en.FindSuggestionsFilter", func() patterns.RuleFilter {
		return NewFindSuggestionsFilter()
	})
	// Partial POS on a regexp-extracted substring; fail-closed without english.dict tagger.
	c.Register("org.languagetool.rules.en.NoDisambiguationEnglishPartialPosTagFilter", func() patterns.RuleFilter {
		return NewNoDisambiguationEnglishPartialPosTagFilter(nil)
	})
	// EnglishPartialPosTagFilter needs tagger+disambiguator; fail-closed until both are wired.
	c.Register("org.languagetool.rules.en.EnglishPartialPosTagFilter", func() patterns.RuleFilter {
		return NewEnglishPartialPosTagFilter(nil)
	})
	// AdvancedSynthesizer: empty subclass; fail-closed without WireDefaultSynthesize.
	c.Register("org.languagetool.rules.en.AdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return NewAdvancedSynthesizerFilter()
	})
	c.Register("org.languagetool.rules.en.EnglishConvertToSentenceCaseFilter", func() patterns.RuleFilter {
		// Embeds ConvertToSentenceCaseFilter; promotes AcceptRuleMatch (exception: "me").
		return NewEnglishConvertToSentenceCaseFilter()
	})
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



