package filters

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Built-in masdar→verb lemma map (subset of Java ArabicMasdarToVerbFilter).
var defaultMasdar2Verb = map[string][]string{
	"عمل":   {"عَمِلَ"},
	"إعمال": {"أَعْمَلَ"},
	"تعميل": {"عَمَّلَ"},
	"ضرب":   {"ضَرَبَ"},
	"أكل":   {"أَكَلَ"},
	"إجابة": {"أَجَابَ"},
}

// Authorized auxiliary lemmas for قام-style constructions.
var authorizeAuxLemma = map[string]struct{}{
	"قَامَ": {},
	"قام":  {},
}

// ArabicMasdarToVerbFilter ports suggestion generation for masdar→verb patterns.
type ArabicMasdarToVerbFilter struct {
	Masdar2Verb map[string][]string
}

func NewArabicMasdarToVerbFilter() *ArabicMasdarToVerbFilter {
	m := map[string][]string{}
	for k, v := range defaultMasdar2Verb {
		cp := make([]string, len(v))
		copy(cp, v)
		m[k] = cp
	}
	return &ArabicMasdarToVerbFilter{Masdar2Verb: m}
}

// LoadMasdarMap merges path-style replace data (lemma → replacements).
func (f *ArabicMasdarToVerbFilter) LoadMasdarMap(data map[string][]string) {
	if f.Masdar2Verb == nil {
		f.Masdar2Verb = map[string][]string{}
	}
	for k, v := range data {
		f.Masdar2Verb[k] = append([]string{}, v...)
	}
}

// FilterAuxLemmas keeps only authorized auxiliary lemmas.
func FilterAuxLemmas(lemmas []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, lem := range lemmas {
		ok := false
		if _, hit := authorizeAuxLemma[lem]; hit {
			ok = true
		} else if _, hit := authorizeAuxLemma[tools.RemoveTashkeel(lem)]; hit {
			ok = true
		}
		if !ok {
			continue
		}
		if _, dup := seen[lem]; dup {
			continue
		}
		seen[lem] = struct{}{}
		out = append(out, lem)
	}
	return out
}

// SuggestVerbsForMasdar returns verb lemmas for a masdar lemma.
func (f *ArabicMasdarToVerbFilter) SuggestVerbsForMasdar(masdarLemma string) []string {
	if f == nil {
		return nil
	}
	if v, ok := f.Masdar2Verb[masdarLemma]; ok {
		return append([]string{}, v...)
	}
	// try without tashkeel
	if v, ok := f.Masdar2Verb[tools.RemoveTashkeel(masdarLemma)]; ok {
		return append([]string{}, v...)
	}
	return nil
}

// AcceptArgs builds replacements when args contain noun (masdar) and optional verb aux.
// Returns a shallow copy of match with suggestions (caller may attach).
func (f *ArabicMasdarToVerbFilter) SuggestionsFromArgs(args map[string]string) []string {
	masdar := args["noun"]
	if masdar == "" {
		masdar = args["masdar"]
	}
	if masdar == "" {
		return nil
	}
	// strip definite article-ish prefix for lookup
	key := strings.TrimPrefix(tools.RemoveTashkeel(masdar), "ال")
	verbs := f.SuggestVerbsForMasdar(key)
	if len(verbs) == 0 {
		verbs = f.SuggestVerbsForMasdar(masdar)
	}
	// dedupe
	seen := map[string]struct{}{}
	var out []string
	for _, v := range verbs {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// ApplySuggestions appends suggestions onto the match (nil-safe).
func ApplySuggestions(match *rules.RuleMatch, suggestions []string) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	if len(suggestions) == 0 {
		return match
	}
	cur := match.GetSuggestedReplacements()
	match.SetSuggestedReplacements(append(append([]string{}, cur...), suggestions...))
	return match
}

// AcceptRuleMatch ports ArabicMasdarToVerbFilter.acceptRuleMatch (surface masdar map path).
// Full Java path also filters aux lemmas via tagger and inflects via synthesizer; without those
// hooks we only emit mapped verb lemmas (fail-closed empty when map misses).
func (f *ArabicMasdarToVerbFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	sugs := f.SuggestionsFromArgs(arguments)
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	if len(sugs) > 0 {
		out.SetSuggestedReplacements(sugs)
	}
	return out
}
