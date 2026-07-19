package filters

import (
	"fmt"
	"strconv"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// adjTagMgr used for Java getLemmas(..., "adj") POS filter (isAdj → starts with NA).
var adjTagMgr = ar_tag.NewArabicTagManager()

// ArabicAdjectiveToExclamationFilter ports org.languagetool.rules.ar.filters.ArabicAdjectiveToExclamationFilter.
// Adj2Comp from official /ar/arabic_adjective_exclamation.txt (Java loadFromPath).
type ArabicAdjectiveToExclamationFilter struct {
	Adj2Comp map[string][]string
}

func NewArabicAdjectiveToExclamationFilter() *ArabicAdjectiveToExclamationFilter {
	return &ArabicAdjectiveToExclamationFilter{Adj2Comp: loadOfficialAdjExclamationMap()}
}

// AcceptRuleMatch ports ArabicAdjectiveToExclamationFilter.acceptRuleMatch.
// Args: adj, noun, adj_pos (1-based pattern token index for the adjective).
// Java: tagger.getLemmas(patternTokens[adj_pos], "adj") only — no surface invent.
func (f *ArabicAdjectiveToExclamationFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	noun := arguments["noun"]
	adjPosStr := arguments["adj_pos"]
	adjTokenIndex, err := strconv.Atoi(adjPosStr)
	if err != nil {
		panic(fmt.Sprintf("Error parsing adj_pos from : %s", adjPosStr))
	}
	adjTokenIndex-- // 1-based → 0-based

	// Java: tagger.getLemmas(..., "adj") — reading lemmas only (no surface / args invent).
	var adjLemmas []string
	seen := map[string]struct{}{}
	addLemma := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		adjLemmas = append(adjLemmas, s)
	}
	if adjTokenIndex >= 0 && adjTokenIndex < len(patternTokens) && patternTokens[adjTokenIndex] != nil {
		tok := patternTokens[adjTokenIndex]
		for _, r := range tok.GetReadings() {
			if r == nil || r.GetLemma() == nil {
				continue
			}
			// Java: tagmanager.isAdj(postag) && type.equals("adj")
			pos := r.GetPOSTag()
			if pos == nil || !adjTagMgr.IsAdj(*pos) {
				continue
			}
			addLemma(*r.GetLemma())
		}
	}

	var compList []string
	compSeen := map[string]struct{}{}
	for _, lem := range adjLemmas {
		for _, c := range f.ComparativesFor(lem) {
			if _, ok := compSeen[c]; ok {
				continue
			}
			compSeen[c] = struct{}{}
			compList = append(compList, c)
		}
	}

	sugs := PrepareExclamationSuggestionsList(compList, noun)
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	if len(sugs) > 0 {
		out.SetSuggestedReplacements(sugs)
	}
	return out
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

// PrepareExclamationSuggestions ports prepareSuggestions(comp, noun).
func PrepareExclamationSuggestions(comp, noun string) []string {
	if comp == "" {
		return nil
	}
	b := comp
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

// PrepareExclamationSuggestionsList maps each comparative.
func PrepareExclamationSuggestionsList(compList []string, noun string) []string {
	var out []string
	for _, c := range compList {
		out = append(out, PrepareExclamationSuggestions(c, noun)...)
	}
	return out
}

func isArabicPronoun(noun string) bool {
	// Java isPronoun list (plus tools map for attached forms used in tests).
	switch noun {
	case "هو", "هي", "هم", "هما", "أنا", "هن", "نحن":
		return true
	}
	_, ok := tools.IsolatedToAttachedPronoun[noun]
	return ok
}

func endsWithBSpace(s string) bool {
	return len(s) >= 2 && s[len(s)-2:] == " ب"
}
