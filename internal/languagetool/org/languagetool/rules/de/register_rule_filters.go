package de

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

func init() {
	// Best-effort tagger for POS-aware filters (no-op without resources).
	WireGermanTaggerDefaults()
	c := patterns.GlobalRuleFilterCreator
	c.Register("org.languagetool.rules.de.FutureDateFilter", func() patterns.RuleFilter {
		return NewFutureDateFilter()
	})
	c.Register("org.languagetool.rules.de.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
	c.Register("org.languagetool.rules.de.YMDDateCheckFilter", func() patterns.RuleFilter {
		return NewYMDDateCheckFilter()
	})
	c.Register("org.languagetool.rules.de.NewYearDateFilter", func() patterns.RuleFilter {
		return NewNewYearDateFilter()
	})
	c.Register("org.languagetool.rules.de.YMDNewYearDateFilter", func() patterns.RuleFilter {
		return NewYMDNewYearDateFilter()
	})
	c.Register("org.languagetool.rules.de.RemoveUnknownCompoundsFilter", func() patterns.RuleFilter {
		return NewRemoveUnknownCompoundsFilter()
	})
	c.Register("org.languagetool.rules.de.PotentialCompoundFilter", func() patterns.RuleFilter {
		return NewPotentialCompoundFilter()
	})
	c.Register("org.languagetool.rules.de.CompoundCheckFilter", func() patterns.RuleFilter {
		return NewCompoundCheckFilter()
	})
	c.Register("org.languagetool.rules.de.InsertCommaFilter", func() patterns.RuleFilter {
		return NewInsertCommaFilter()
	})
	c.Register("org.languagetool.rules.de.RecentYearFilter", func() patterns.RuleFilter {
		return NewRecentYearFilter()
	})
	c.Register("org.languagetool.rules.de.ValidWordFilter", func() patterns.RuleFilter {
		return NewValidWordFilter()
	})
	c.Register("org.languagetool.rules.de.GermanSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return NewGermanSuppressMisspelledSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.de.UppercaseNounReadingFilter", func() patterns.RuleFilter {
		return NewUppercaseNounReadingFilter()
	})
	c.Register("org.languagetool.rules.de.GermanNumberInWordFilter", func() patterns.RuleFilter {
		return NewGermanNumberInWordFilter()
	})
	c.Register("org.languagetool.rules.de.AdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return NewAdvancedSynthesizerFilter()
	})
	// AdaptSuggestionFilter: GenderOf+Synthesize from discovered DE resources when present.
	c.Register("org.languagetool.rules.de.AdaptSuggestionFilter", func() patterns.RuleFilter {
		return WireAdaptSuggestionFilter()
	})
}

func deFutureDateCore() *rules.FutureDateFilterCore {
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

func deDateCheckWithSuggestions() *rules.AbstractDateCheckWithSuggestionsFilter {
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
			FormatDayOfWeek: func(t time.Time) string {
				// German long weekday names (DateFilterHelper style)
				names := []string{"Sonntag", "Montag", "Dienstag", "Mittwoch", "Donnerstag", "Freitag", "Samstag"}
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
		// Java DateCheckFilter.getErrorMessageWrongYear
		ErrorMessageWrongYear: `Dieses Datum stimmt nicht mit dem Tag überein. Meinten Sie "{currentYear}"?`,
	}
}

func deNewYearDateCore() *rules.NewYearDateFilterCore {
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
