package filters

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicAdjectiveExclamationFileName ports ArabicAdjectiveToExclamationFilter.FILE_NAME.
const ArabicAdjectiveExclamationFileName = "/ar/arabic_adjective_exclamation.txt"

// adjTagMgr used for Java ArabicTagger.getLemmas(..., "adj") POS filter (isAdj → starts with NA).
var adjTagMgr = ar_tag.NewArabicTagManager()

// ArabicAdjectiveToExclamationFilter ports org.languagetool.rules.ar.filters.ArabicAdjectiveToExclamationFilter.
// Adj2Comp from official /ar/arabic_adjective_exclamation.txt (Java loadFromPath / SimpleReplaceDataLoader).
//
// Note: Java field adj2comp (3 put() rows) is dead unused on the accept path — not used as production data.
// Note: Java private getAttachedPronoun is dead; production uses ArabicWordMaps.getAttachedPronoun.
type ArabicAdjectiveToExclamationFilter struct {
	// Adj2Comp ports adj2compList (file-backed). Keys match SimpleReplaceDataLoader (no invent trim).
	Adj2Comp map[string][]string
}

// NewArabicAdjectiveToExclamationFilter ports the Java ctor: adj2compList = loadFromPath(FILE_NAME).
func NewArabicAdjectiveToExclamationFilter() *ArabicAdjectiveToExclamationFilter {
	return &ArabicAdjectiveToExclamationFilter{Adj2Comp: loadOfficialAdjExclamationMap()}
}

// GetDataFilePath ports ArabicAdjectiveToExclamationFilter.getDataFilePath.
func GetDataFilePath() string {
	return ArabicAdjectiveExclamationFileName
}

// AcceptRuleMatch ports ArabicAdjectiveToExclamationFilter.acceptRuleMatch.
// Args: adj (extracted unused after), noun, adj_pos (1-based; parseInt-1; NFE → panic with Java message).
// Java: tagger.getLemmas(patternTokens[adjTokenIndex], "adj") only — no surface invent.
func (f *ArabicAdjectiveToExclamationFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	// Java: arguments.get("adj") extracted then unused
	_ = arguments["adj"]
	noun := arguments["noun"]
	adjPosStr := arguments["adj_pos"]
	adjTokenIndex, err := strconv.Atoi(adjPosStr)
	if err != nil {
		// Java: throw new RuntimeException("Error parsing adj_pos from : " + arguments.get("adj_pos"), e);
		panic(fmt.Sprintf("Error parsing adj_pos from : %s", adjPosStr))
	}
	adjTokenIndex-- // 1-based → 0-based

	// Java: tagger.getLemmas(patternTokens[adjTokenIndex], "adj") — no bounds check (AIOOBE if bad).
	adjLemmas := adjLemmasFromToken(patternTokens[adjTokenIndex])

	// Java: for each lemma adj2compList.get(lemma); HashSet dedupe (order not guaranteed).
	// Go preserves first-seen insertion order for determinism; multiset matches Java.
	var compList []string
	compSeen := map[string]struct{}{}
	for _, lem := range adjLemmas {
		for _, c := range f.comparativesExact(lem) {
			if _, ok := compSeen[c]; ok {
				continue
			}
			compSeen[c] = struct{}{}
			compList = append(compList, c)
		}
	}

	// Java: new RuleMatch(... message, shortMessage); addSuggestedReplacement each.
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	sugs := PrepareExclamationSuggestionsList(compList, noun)
	if len(sugs) > 0 {
		out.SetSuggestedReplacements(sugs)
	}
	return out
}

// ComparativesFor is a test/helper exact map lookup (Java adj2compList.get).
// No RemoveTashkeel invent — Java only exact get(adjLemma).
func (f *ArabicAdjectiveToExclamationFilter) ComparativesFor(adjLemma string) []string {
	return f.comparativesExact(adjLemma)
}

func (f *ArabicAdjectiveToExclamationFilter) comparativesExact(adjLemma string) []string {
	if f == nil || f.Adj2Comp == nil {
		return nil
	}
	if v, ok := f.Adj2Comp[adjLemma]; ok {
		return append([]string{}, v...)
	}
	return nil
}

// adjLemmasFromToken ports ArabicTagger.getLemmas(readings, "adj").
func adjLemmasFromToken(tok *languagetool.AnalyzedTokenReadings) []string {
	if tok == nil {
		return nil
	}
	var lemmaList []string
	seen := map[string]struct{}{}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		pos := r.GetPOSTag()
		if pos == nil || !adjTagMgr.IsAdj(*pos) {
			continue
		}
		lem := ""
		if r.GetLemma() != nil {
			lem = *r.GetLemma()
		}
		if _, ok := seen[lem]; ok {
			continue
		}
		seen[lem] = struct{}{}
		lemmaList = append(lemmaList, lem)
	}
	return lemmaList
}

// PrepareExclamationSuggestionsList ports prepareSuggestions(List, noun).
func PrepareExclamationSuggestionsList(compList []string, noun string) []string {
	var sugList []string
	for _, comp := range compList {
		sugList = append(sugList, PrepareExclamationSuggestions(comp, noun)...)
	}
	return sugList
}

// PrepareExclamationSuggestions ports prepareSuggestions(comp, noun).
// Returns a singleton list with the built suggestion string (Java always adds one).
func PrepareExclamationSuggestions(comp, noun string) []string {
	// Java starts with comp even if empty; no early empty-comp skip.
	var b strings.Builder
	b.WriteString(comp)
	if noun == "" {
		// Java: noun == null || noun.isEmpty() → leave as comp
	} else if isArabicPronoun(noun) {
		// no space; ArabicWordMaps.getAttachedPronoun (not the dead private map)
		b.WriteString(tools.GetAttachedPronoun(noun))
	} else {
		// if comparative is of second form don't add a space
		if !strings.HasSuffix(comp, " ب") {
			b.WriteString(" ")
		}
		b.WriteString(noun)
	}
	return []string{b.String()}
}

// isArabicPronoun ports ArabicAdjectiveToExclamationFilter.isPronoun (exact list only).
// Java: هو/هي/هم/هما/أنا — does not include هن/نحن (those exist only on the dead private map).
func isArabicPronoun(word string) bool {
	if word == "" {
		return false
	}
	switch word {
	case "هو", "هي", "هم", "هما", "أنا":
		return true
	default:
		return false
	}
}
