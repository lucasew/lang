package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FindSuggestionsFilter ports org.languagetool.rules.ca.FindSuggestionsFilter
// (extends AbstractFindSuggestionsFilter with MorfologikSpeller + CatalanTagger).
//
// Default SpellingSuggestions uses process-wide WireCatalanFilterSpeller /
// FilterDictSuggest (Java: speller.findSimilarWords; resource /ca/ca-ES_spelling.dict).
// Without a dict, Accept fails closed unless SetSpellingSuggestions overrides.
// Tag: CatalanTagger.INSTANCE_CAT via process-wide FilterTagWord.
// Java getSynthesizer() override is commented out → null (no replacements2 synth path).
type FindSuggestionsFilter struct {
	*rules.AbstractFindSuggestionsFilter
	// spellingOverride true when host/tests set SpellingSuggestions via SetSpellingSuggestions.
	spellingOverride bool
}

// LemmasToIgnore / LemmasToAllow port CA FindSuggestionsFilter static arrays.
var (
	LemmasToIgnore = []string{"enterar", "sentar", "conseguir", "alcançar", "liar", "vore"}
	LemmasToAllow  = []string{"enter", "sentir"}
	// ELA_GEMINADA: l .·•∙ etc. l → l·l
	elaGeminada = regexp.MustCompile(`(?i)(l)[\.\x{2022}\x{22C5}\x{2219}\x{F0D7}\-](l)`)
)

func NewFindSuggestionsFilter() *FindSuggestionsFilter {
	f := &FindSuggestionsFilter{
		AbstractFindSuggestionsFilter: &rules.AbstractFindSuggestionsFilter{
			// Java: speller.findSimilarWords(atr.getToken())
			SpellingSuggestions: defaultCASpellingSuggestions,
			// Java: getTagger() → CatalanTagger.INSTANCE_CAT
			Tag: FilterTagWord,
			// Java: // getSynthesizer() commented out — CatalanSynthesizer not used
		},
	}
	f.IsSuggestionException = caIsSuggestionException
	f.PreProcessWrongWord = caPreProcessWrongWord
	return f
}

func defaultCASpellingSuggestions(atr *languagetool.AnalyzedTokenReadings) []string {
	if atr == nil || !FilterDictAvailable() {
		return nil
	}
	return FilterDictSuggest(atr.GetToken())
}

func caIsSuggestionException(atr *languagetool.AnalyzedTokenReadings) bool {
	if atr == nil {
		return false
	}
	// hasAnyLemma(LemmasToIgnore) && !hasAnyLemma(LemmasToAllow)
	if !atr.HasAnyLemma(LemmasToIgnore...) {
		return false
	}
	return !atr.HasAnyLemma(LemmasToAllow...)
}

func caPreProcessWrongWord(word string) string {
	word = strings.ReplaceAll(word, " ", "")
	return elaGeminada.ReplaceAllString(word, "$1·$2")
}

// SetSpellingSuggestions installs a host/test findSimilarWords-style hook.
func (f *FindSuggestionsFilter) SetSpellingSuggestions(fn func(atr *languagetool.AnalyzedTokenReadings) []string) {
	if f == nil || f.AbstractFindSuggestionsFilter == nil || fn == nil {
		return
	}
	f.spellingOverride = true
	f.SpellingSuggestions = fn
}

// AcceptRuleMatch ports AbstractFindSuggestionsFilter.acceptRuleMatch with CA fail-closed gate.
func (f *FindSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || f.AbstractFindSuggestionsFilter == nil || match == nil {
		return nil
	}
	// Java constructs MorfologikSpeller when resource exists; without dict / override, fail-closed.
	if !f.spellingOverride && !FilterDictAvailable() {
		return nil
	}
	return f.AbstractFindSuggestionsFilter.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
