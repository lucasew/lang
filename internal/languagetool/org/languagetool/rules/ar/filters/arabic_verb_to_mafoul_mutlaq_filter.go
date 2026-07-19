package filters

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_synth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicVerbToMafoulMutlaqFilter ports org.languagetool.rules.ar.filters.ArabicVerbToMafoulMutlaqFilter.
// Java always uses ArabicSynthesizer.inflectMafoulMutlq / inflectAdjectiveTanwinNasb (static morph).
// Verb lemmas come from pattern token readings (Java tagger.getLemmas verb) — no surface invent.
// Verb2Masdar from official /ar/arabic_verb_masdar.txt (Java loadFromPath).
type ArabicVerbToMafoulMutlaqFilter struct {
	Verb2Masdar map[string][]string
	// InflectMafoul optional override (default: ar_synth.InflectMafoulMutlq).
	InflectMafoul func(word string) string
	// InflectAdj optional override (default: ar_synth.InflectAdjectiveTanwinNasb).
	InflectAdj func(word string, feminin bool) string
}

func NewArabicVerbToMafoulMutlaqFilter() *ArabicVerbToMafoulMutlaqFilter {
	return &ArabicVerbToMafoulMutlaqFilter{Verb2Masdar: loadOfficialVerbMasdarMap()}
}

func (f *ArabicVerbToMafoulMutlaqFilter) inflectMafoul(word string) string {
	if f != nil && f.InflectMafoul != nil {
		return f.InflectMafoul(word)
	}
	return ar_synth.InflectMafoulMutlq(word)
}

func (f *ArabicVerbToMafoulMutlaqFilter) inflectAdj(word string, feminin bool) string {
	if f != nil && f.InflectAdj != nil {
		return f.InflectAdj(word, feminin)
	}
	return ar_synth.InflectAdjectiveTanwinNasb(word, feminin)
}

// AcceptRuleMatch ports ArabicVerbToMafoulMutlaqFilter.acceptRuleMatch.
// Args: verb, adj. Verb lemmas from patternTokens[0] readings only (Java tagger path).
func (f *ArabicVerbToMafoulMutlaqFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	verb := arguments["verb"]
	adj := arguments["adj"]

	// Java: tagger.getLemmas(patternTokens[0], "verb") — lemma readings only, no surface invent.
	var verbLemmas []string
	seen := map[string]struct{}{}
	addLemma := func(s string) {
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
			if r == nil || r.GetLemma() == nil {
				continue
			}
			// Prefer verb-tagged readings when POS present; untagged lemma still accepted
			// only when POS is empty (injected lemma tests) or starts with V.
			if pos := r.GetPOSTag(); pos != nil && *pos != "" && !strings.HasPrefix(*pos, "V") {
				continue
			}
			addLemma(*r.GetLemma())
		}
	}

	inflectedAdjMasculine := f.inflectAdj(adj, false)
	inflectedAdjFeminin := f.inflectAdj(adj, true)

	var inflectedMasdarList []string
	var inflectedAdjList []string
	for _, lemma := range verbLemmas {
		for _, msdr := range f.MasdarsForVerb(lemma) {
			if msdr == "" {
				continue
			}
			inflectedMasdarList = append(inflectedMasdarList, f.inflectMafoul(msdr))
			if strings.HasSuffix(msdr, string(tools.ArabicTehMarbuta)) {
				inflectedAdjList = append(inflectedAdjList, inflectedAdjFeminin)
			} else {
				inflectedAdjList = append(inflectedAdjList, inflectedAdjMasculine)
			}
		}
	}

	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	sugSeen := map[string]struct{}{}
	var sugs []string
	for i, msdr := range inflectedMasdarList {
		phrase := verb + " " + msdr
		if i < len(inflectedAdjList) && inflectedAdjList[i] != "" {
			phrase += " " + inflectedAdjList[i]
		}
		if _, ok := sugSeen[phrase]; ok {
			continue
		}
		sugSeen[phrase] = struct{}{}
		sugs = append(sugs, phrase)
	}
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

// SuggestMafoulMutlaq builds inflected masdar suggestions (optionally doubled) — test helper.
func (f *ArabicVerbToMafoulMutlaqFilter) SuggestMafoulMutlaq(verbLemma string) []string {
	ms := f.MasdarsForVerb(verbLemma)
	var out []string
	for _, m := range ms {
		inf := f.inflectMafoul(m)
		out = append(out, inf)
		out = append(out, inf+" "+inf)
	}
	return out
}
