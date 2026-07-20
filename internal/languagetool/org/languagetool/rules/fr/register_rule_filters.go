package fr

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	// Suppress-misspelled: FilterDictIsMisspelled when french.dict wired (Java default spelling rule).
	c.Register("org.languagetool.rules.fr.FrenchSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return NewFrenchSuppressMisspelledSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.fr.MakeContractionsFilter", func() patterns.RuleFilter {
		return NewMakeContractionsFilter()
	})
	c.Register("org.languagetool.rules.fr.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
	c.Register("org.languagetool.rules.fr.NewYearDateFilter", func() patterns.RuleFilter {
		return NewNewYearDateFilter()
	})
	c.Register("org.languagetool.rules.fr.DMYDateCheckFilter", func() patterns.RuleFilter {
		return NewDMYDateCheckFilter()
	})
	c.Register("org.languagetool.rules.fr.SuggestionsFilter", func() patterns.RuleFilter {
		return NewSuggestionsFilter()
	})
	// FindSuggestionsFilter: French multi-query spelling via WireFrenchFilterSpeller; Tag/Synthesize optional.
	c.Register("org.languagetool.rules.fr.FindSuggestionsFilter", func() patterns.RuleFilter {
		return NewFindSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.fr.AdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return NewAdvancedSynthesizerFilter()
	})
	// Number-in-word: MorfologikFrenchSpellerRule via WireFrenchFilterSpeller; fail-closed without dict.
	c.Register("org.languagetool.rules.fr.FrenchNumberInWordFilter", func() patterns.RuleFilter {
		return NewFrenchNumberInWordFilter()
	})
	// Interrogative/imperative verb: needs FR speller+tagger; optional synth for je-forms.
	c.Register("org.languagetool.rules.fr.InterrogativeVerbFilter", func() patterns.RuleFilter {
		return NewInterrogativeVerbFilter()
	})
	// WordWithDeterminerFilter: full Accept; wire FrenchSynthesizer via Synthesize.
	c.Register("org.languagetool.rules.fr.WordWithDeterminerFilter", func() patterns.RuleFilter {
		return NewWordWithDeterminerFilter()
	})
	// Full AcceptRuleMatch port; wire FrenchSynthesizer via Synthesize when available.
	c.Register("org.languagetool.rules.fr.PostponedAdjectiveConcordanceFilter", func() patterns.RuleFilter {
		return NewPostponedAdjectiveConcordanceFilter()
	})
	// FrenchPartialPosTagFilter: PartialPosTagFilter + FR tagger+disambiguator (Tag hook; nil → fail-closed).
	c.Register("org.languagetool.rules.fr.FrenchPartialPosTagFilter", func() patterns.RuleFilter {
		return NewFrenchPartialPosTagFilter(nil)
	})
}

func frDateCheckWithSuggestions() *rules.AbstractDateCheckWithSuggestionsFilter {
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
			FormatDayOfWeek: formatFRDayOfWeek,
			GetMonth: func(localized string) int {
				m, err := h.GetMonth(localized)
				if err != nil {
					return 0
				}
				return int(m)
			},
		},
		// Java DateCheckFilter.getErrorMessageWrongYear
		// Java ends with NBSP before '?': "\"{currentYear}\"\u00a0?"
	ErrorMessageWrongYear: "Cette date est incorrecte. Faites-vous référence à l'année \"{currentYear}\"\u00a0?",
	}
}

func frNewYearDateCore() *rules.NewYearDateFilterCore {
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
