package ca

import (
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// Register language RuleFilter class names for grammar/style XML (fail-closed unknowns).
func init() {
	c := patterns.GlobalRuleFilterCreator
	// Suppress-misspelled: Catalan isMisspelled override (null speller → true; chunk + speller).
	c.Register("org.languagetool.rules.ca.CatalanSuppressMisspelledSuggestionsFilter", func() patterns.RuleFilter {
		return NewCatalanSuppressMisspelledSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.ca.DiacriticsCheckFilter", func() patterns.RuleFilter {
		return diacriticsCheckRuleFilter{inner: NewDiacriticsCheckFilter().ConfusionCheckFilter}
	})
	// FindSuggestionsFilter: Morfologik findSimilarWords via WireCatalanFilterSpeller; Tag optional.
	c.Register("org.languagetool.rules.ca.FindSuggestionsFilter", func() patterns.RuleFilter {
		return NewFindSuggestionsFilter()
	})
	// AdvancedSynthesizer: empty subclass; fail-closed without WireDefaultSynthesize.
	c.Register("org.languagetool.rules.ca.AdvancedSynthesizerFilter", func() patterns.RuleFilter {
		return NewAdvancedSynthesizerFilter()
	})
	c.Register("org.languagetool.rules.ca.DateCheckFilter", func() patterns.RuleFilter {
		return NewDateCheckFilter()
	})
	c.Register("org.languagetool.rules.ca.NewYearDateFilter", func() patterns.RuleFilter {
		return NewNewYearDateFilter()
	})
	c.Register("org.languagetool.rules.ca.TextToNumberFilter", func() patterns.RuleFilter {
		return NewTextToNumberFilter()
	})
	// Number-in-word: MorfologikCatalanSpellerRule via WireCatalanFilterSpeller; fail-closed without dict.
	c.Register("org.languagetool.rules.ca.CatalanNumberInWordFilter", func() patterns.RuleFilter {
		return NewCatalanNumberInWordFilter()
	})
	// Number speller needs CatalanSynthesizer.getSpelledNumber; fail-closed without SpellNumber.
	c.Register("org.languagetool.rules.ca.CatalanNumberSpellerFilter", func() patterns.RuleFilter {
		return NewCatalanNumberSpellerFilter(nil)
	})
	c.Register("org.languagetool.rules.ca.CatalanRemoteRewriteFilter", func() patterns.RuleFilter {
		return NewCatalanRemoteRewriteFilter()
	})
	c.Register("org.languagetool.rules.ca.AnarASuggestionsFilter", func() patterns.RuleFilter {
		return NewAnarASuggestionsFilter()
	})
	c.Register("org.languagetool.rules.ca.PortarGerundiSuggestionsFilter", func() patterns.RuleFilter {
		return NewPortarGerundiSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.ca.FindSuggestionsEsFilter", func() patterns.RuleFilter {
		return NewFindSuggestionsEsFilter()
	})
	c.Register("org.languagetool.rules.ca.SynthesizeWithDAFilter", func() patterns.RuleFilter {
		return NewSynthesizeWithDAFilter()
	})
	c.Register("org.languagetool.rules.ca.SynthesizeWithAnyDeterminerFilter", func() patterns.RuleFilter {
		return NewSynthesizeWithAnyDeterminerFilter()
	})
	c.Register("org.languagetool.rules.ca.ConvertToGenderAndNumberFilter", func() patterns.RuleFilter {
		return NewConvertToGenderAndNumberFilter()
	})
	c.Register("org.languagetool.rules.ca.OblidarseSugestionsFilter", func() patterns.RuleFilter {
		return NewOblidarseSugestionsFilter()
	})
	c.Register("org.languagetool.rules.ca.PossessiusRedundantsFilter", func() patterns.RuleFilter {
		return NewPossessiusRedundantsFilter()
	})
	c.Register("org.languagetool.rules.ca.EnNoInfinitiuSuggestionFilter", func() patterns.RuleFilter {
		return NewEnNoInfinitiuSuggestionFilter()
	})
	c.Register("org.languagetool.rules.ca.PortarTempsSuggestionsFilter", func() patterns.RuleFilter {
		return NewPortarTempsSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.ca.DonarTempsSuggestionsFilter", func() patterns.RuleFilter {
		return NewDonarTempsSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.ca.DonarseliBeFilter", func() patterns.RuleFilter {
		return NewDonarseliBeFilter()
	})
	// AdjustVerbSuggestionsFilter: full Accept; needs Synthesize for verb forms.
	c.Register("org.languagetool.rules.ca.AdjustVerbSuggestionsFilter", func() patterns.RuleFilter {
		return NewAdjustVerbSuggestionsFilter()
	})
	c.Register("org.languagetool.rules.ca.AdjustPronounsFilter", func() patterns.RuleFilter {
		return NewAdjustPronounsFilter()
	})
	// Full AcceptRuleMatch port; wire Catalan synthesizer via Synthesize when available.
	c.Register("org.languagetool.rules.ca.PostponedAdjectiveConcordanceFilter", func() patterns.RuleFilter {
		return NewPostponedAdjectiveConcordanceFilter()
	})
}












func caDateCheckWithSuggestions() *rules.AbstractDateCheckWithSuggestionsFilter {
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
			FormatDayOfWeek: formatCADayOfWeek,
			GetMonth: func(localized string) int {
				m, err := h.GetMonth(localized)
				if err != nil {
					return 0
				}
				return int(m)
			},
		},
		// Java DateCheckFilter.getErrorMessageWrongYear
		ErrorMessageWrongYear: `Aquesta data no és correcta. ¿Us referiu a l'any "{currentYear}"?`,
	}
}

func caNewYearDateCore() *rules.NewYearDateFilterCore {
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


type diacriticsCheckRuleFilter struct {
	inner *rules.ConfusionCheckFilter
}

func (f diacriticsCheckRuleFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, pos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments, pos, patternTokens, tokenPositions)
}
