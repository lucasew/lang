package ca

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// FindSuggestionsFilter ports org.languagetool.rules.ca.FindSuggestionsFilter
// (extends AbstractFindSuggestionsFilter with MorfologikSpeller + CatalanTagger).
//
// Default SpellingSuggestions uses process-wide WireCatalanFilterSpeller /
// FilterDictSuggest (Java: speller.findSimilarWords; resource /ca/ca-ES_spelling.dict).
// Without a dict, Accept fails closed unless SetSpellingSuggestions overrides.
// Tag / Synthesize: CatalanTagger + CatalanSynthesizer via process-wide hooks.
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
			// Java: getTagger() → CatalanTagger
			Tag: FilterTagWord,
			// Java: getSynthesizer().synthesize(at, desiredPostag, true)
			Synthesize: catalanFindSuggestionsSynthesize,
		},
	}
	f.IsSuggestionException = caIsSuggestionException
	f.PreProcessWrongWord = caPreProcessWrongWord
	return f
}

func catalanFindSuggestionsSynthesize(tok *languagetool.AnalyzedToken, postagRE string) []string {
	// Catalan variants use full codes (ca-ES); registry may hold "ca" or "ca-ES".
	for _, code := range []string{"ca", "ca-ES", "ca-ES-valencia"} {
		s := patterns.LanguageSynthesizer(code)
		if s == nil {
			continue
		}
		forms, err := s.SynthesizeRE(tok, postagRE, true)
		if err != nil || len(forms) == 0 {
			continue
		}
		return forms
	}
	return nil
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
