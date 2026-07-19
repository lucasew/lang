package filters

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Built-in masdar→verb lemma map (subset of Java ArabicMasdarToVerbFilter).
// Full list also loads from /ar/arabic_masdar_verb.txt when wired.
var defaultMasdar2Verb = map[string][]string{
	"عمل":   {"عَمِلَ"},
	"إعمال": {"أَعْمَلَ"},
	"تعميل": {"عَمَّلَ"},
	"ضرب":   {"ضَرَبَ"},
	"أكل":   {"أَكَلَ"},
	"إجابة": {"أَجَابَ"},
}

// Authorized auxiliary lemmas for قام-style constructions (Java authorizeLemma).
var authorizeAuxLemma = map[string]struct{}{
	"قَامَ": {},
}

var masdarTagMgr = ar_tag.NewArabicTagManager()

// ArabicMasdarToVerbFilter ports org.languagetool.rules.ar.filters.ArabicMasdarToVerbFilter.
// Accept uses patternTokens[0] verb lemmas + patternTokens[1] masdar lemmas (Java tagger),
// then synthesizer.inflectLemmaLike — no surface invent of verb forms.
type ArabicMasdarToVerbFilter struct {
	Masdar2Verb map[string][]string
	// InflectLemmaLike ports ArabicSynthesizer.inflectLemmaLike.
	// Nil → fail closed (empty suggestions; Java always has synth).
	InflectLemmaLike func(targetLemma string, source *languagetool.AnalyzedToken) []string
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

// FilterAuxLemmas keeps only authorized auxiliary lemmas (Java filterLemmas).
func FilterAuxLemmas(lemmas []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, lem := range lemmas {
		if _, hit := authorizeAuxLemma[lem]; !hit {
			// Java authorizeLemma is exact diacritic match only (قَامَ).
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

// SuggestVerbsForMasdar returns verb lemmas for a masdar lemma (map lookup helper).
func (f *ArabicMasdarToVerbFilter) SuggestVerbsForMasdar(masdarLemma string) []string {
	if f == nil {
		return nil
	}
	if v, ok := f.Masdar2Verb[masdarLemma]; ok {
		return append([]string{}, v...)
	}
	if v, ok := f.Masdar2Verb[tools.RemoveTashkeel(masdarLemma)]; ok {
		return append([]string{}, v...)
	}
	return nil
}

// SuggestionsFromArgs is a map-lookup helper for tests (not the full Java Accept path).
// Prefer AcceptRuleMatch with tagged tokens + InflectLemmaLike.
func (f *ArabicMasdarToVerbFilter) SuggestionsFromArgs(args map[string]string) []string {
	masdar := args["noun"]
	if masdar == "" {
		masdar = args["masdar"]
	}
	if masdar == "" {
		return nil
	}
	key := strings.TrimPrefix(tools.RemoveTashkeel(masdar), "ال")
	verbs := f.SuggestVerbsForMasdar(key)
	if len(verbs) == 0 {
		verbs = f.SuggestVerbsForMasdar(masdar)
	}
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

// lemmasOfType ports ArabicTagger.getLemmas(..., type) for verb/masdar.
func lemmasOfType(tok *languagetool.AnalyzedTokenReadings, kind string) []string {
	if tok == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetLemma() == nil || r.GetPOSTag() == nil {
			continue
		}
		pos, lem := *r.GetPOSTag(), *r.GetLemma()
		ok := false
		switch kind {
		case "verb":
			ok = masdarTagMgr.IsVerb(pos)
		case "masdar":
			ok = masdarTagMgr.IsMasdar(pos)
		}
		if !ok || lem == "" {
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

// AcceptRuleMatch ports ArabicMasdarToVerbFilter.acceptRuleMatch.
// patternTokens[0]=aux verb, [1]=masdar noun. Requires InflectLemmaLike for forms.
func (f *ArabicMasdarToVerbFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	if len(patternTokens) < 2 || patternTokens[0] == nil || patternTokens[1] == nil {
		return out
	}
	if f.InflectLemmaLike == nil {
		// Java always has ArabicSynthesizer; fail closed without inflectLemmaLike.
		return out
	}

	auxVerbLemmas := FilterAuxLemmas(lemmasOfType(patternTokens[0], "verb"))
	if len(auxVerbLemmas) == 0 {
		return out
	}
	auxSet := map[string]struct{}{}
	for _, l := range auxVerbLemmas {
		auxSet[l] = struct{}{}
	}
	masdarLemmas := lemmasOfType(patternTokens[1], "masdar")

	seen := map[string]struct{}{}
	var verbList []string
	// Java: for each aux reading with authorized lemma × masdar lemmas × mapped verbs
	for _, auxTok := range patternTokens[0].GetReadings() {
		if auxTok == nil || auxTok.GetLemma() == nil {
			continue
		}
		if _, ok := auxSet[*auxTok.GetLemma()]; !ok {
			continue
		}
		for _, masdarLemma := range masdarLemmas {
			for _, vrbLem := range f.SuggestVerbsForMasdar(masdarLemma) {
				for _, form := range f.InflectLemmaLike(vrbLem, auxTok) {
					if form == "" {
						continue
					}
					if _, dup := seen[form]; dup {
						continue
					}
					seen[form] = struct{}{}
					verbList = append(verbList, form)
				}
			}
		}
	}
	if len(verbList) > 0 {
		out.SetSuggestedReplacements(verbList)
	}
	// arguments verb/noun are Java extract only; unused for suggestions once tokens are tagged.
	_ = arguments
	return out
}
