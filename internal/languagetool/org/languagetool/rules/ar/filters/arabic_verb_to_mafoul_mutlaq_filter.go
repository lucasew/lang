package filters

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// Default verb→masdar map (Java ArabicVerbToMafoulMutlaqFilter built-in subset).
// Full list also loads from /ar/arabic_verb_masdar.txt when wired.
var defaultVerb2Masdar = map[string][]string{
	"عَمِلَ":  {"عمل"},
	"أَعْمَلَ": {"إعمال"},
	"عَمَّلَ":  {"تعميل"},
	"عمل":   {"عمل"},
	"أَكَلَ":  {"أكل"},
	"سَأَلَ":  {"سؤال"},
	"أَجَابَ":  {"إجابة"},
}

// ArabicVerbToMafoulMutlaqFilter ports org.languagetool.rules.ar.filters.ArabicVerbToMafoulMutlaqFilter.
// Full Java path inflects masdar/adj via synthesizer (tanwin nasb); without synth we use surface forms.
type ArabicVerbToMafoulMutlaqFilter struct {
	Verb2Masdar map[string][]string
}

func NewArabicVerbToMafoulMutlaqFilter() *ArabicVerbToMafoulMutlaqFilter {
	m := map[string][]string{}
	for k, v := range defaultVerb2Masdar {
		m[k] = append([]string{}, v...)
	}
	return &ArabicVerbToMafoulMutlaqFilter{Verb2Masdar: m}
}

// AcceptRuleMatch ports ArabicVerbToMafoulMutlaqFilter.acceptRuleMatch.
// Args: verb, adj. Lemmas from patternTokens[0] readings when present.
// Incomplete without synthesizer.inflectMafoulMutlq / inflectAdjectiveTanwinNasb (surface masdar+adj).
func (f *ArabicVerbToMafoulMutlaqFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	verb := arguments["verb"]
	adj := arguments["adj"]

	var verbLemmas []string
	seen := map[string]struct{}{}
	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		verbLemmas = append(verbLemmas, s)
	}
	if len(patternTokens) > 0 && patternTokens[0] != nil {
		for _, r := range patternTokens[0].GetReadings() {
			if r == nil {
				continue
			}
			if r.GetLemma() != nil {
				add(*r.GetLemma())
			}
		}
		add(patternTokens[0].GetToken())
	}
	add(verb)

	var sugs []string
	sugSeen := map[string]struct{}{}
	for _, lemma := range verbLemmas {
		for _, msdr := range f.MasdarsForVerb(lemma) {
			// Java: verb + " " + inflectedMasdar + " " + inflectedAdj
			// Surface path until synthesizer hooks are available.
			phrase := verb + " " + msdr
			if adj != "" {
				phrase += " " + adj
			}
			if _, ok := sugSeen[phrase]; ok {
				continue
			}
			sugSeen[phrase] = struct{}{}
			sugs = append(sugs, phrase)
		}
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	if len(sugs) > 0 {
		out.SetSuggestedReplacements(sugs)
	}
	return out
}

// MasdarsForVerb returns masdar forms for a verb lemma.
func (f *ArabicVerbToMafoulMutlaqFilter) MasdarsForVerb(verbLemma string) []string {
	if f == nil {
		return nil
	}
	if v, ok := f.Verb2Masdar[verbLemma]; ok {
		return append([]string{}, v...)
	}
	if v, ok := f.Verb2Masdar[tools.RemoveTashkeel(verbLemma)]; ok {
		return append([]string{}, v...)
	}
	return nil
}

// SuggestMafoulMutlaq builds "masdar" suggestions (optionally doubled) — helper for tests.
func (f *ArabicVerbToMafoulMutlaqFilter) SuggestMafoulMutlaq(verbLemma string) []string {
	ms := f.MasdarsForVerb(verbLemma)
	var out []string
	for _, m := range ms {
		out = append(out, m)
		out = append(out, m+" "+m)
	}
	return out
}
