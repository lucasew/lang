package filters

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Default adjective → comparative map (Java ArabicAdjectiveToExclamationFilter subset).
var defaultAdj2Comp = map[string][]string{
	"رشيد": {"أرشد"},
	"طويل": {"أطول"},
	"بديع": {"أبدع"},
}

// ArabicAdjectiveToExclamationFilter ports comparative suggestion helpers.
type ArabicAdjectiveToExclamationFilter struct {
	Adj2Comp map[string][]string
}

func NewArabicAdjectiveToExclamationFilter() *ArabicAdjectiveToExclamationFilter {
	m := map[string][]string{}
	for k, v := range defaultAdj2Comp {
		m[k] = append([]string{}, v...)
	}
	return &ArabicAdjectiveToExclamationFilter{Adj2Comp: m}
}

func (f *ArabicAdjectiveToExclamationFilter) ComparativesFor(adjLemma string) []string {
	if f == nil {
		return nil
	}
	if v, ok := f.Adj2Comp[adjLemma]; ok {
		return append([]string{}, v...)
	}
	if v, ok := f.Adj2Comp[tools.RemoveTashkeel(adjLemma)]; ok {
		return append([]string{}, v...)
	}
	return nil
}

// PrepareSuggestions ports prepareSuggestions(comp, noun).
func PrepareExclamationSuggestions(comp, noun string) []string {
	if comp == "" {
		return nil
	}
	var b string
	b = comp
	if noun == "" {
		return []string{b}
	}
	if isArabicPronoun(noun) {
		b += tools.GetAttachedPronoun(noun)
		return []string{b}
	}
	if !endsWithBSpace(comp) {
		b += " "
	}
	b += noun
	return []string{b}
}

// PrepareSuggestionsList maps each comparative.
func PrepareExclamationSuggestionsList(compList []string, noun string) []string {
	var out []string
	for _, c := range compList {
		out = append(out, PrepareExclamationSuggestions(c, noun)...)
	}
	return out
}

func isArabicPronoun(noun string) bool {
	_, ok := tools.IsolatedToAttachedPronoun[noun]
	return ok
}

func endsWithBSpace(s string) bool {
	return len(s) >= 2 && s[len(s)-2:] == " ب"
}
