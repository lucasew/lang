package filters

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_synth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ArabicVerbMasdarFileName ports ArabicVerbToMafoulMutlaqFilter.FILE_NAME.
const ArabicVerbMasdarFileName = "/ar/arabic_verb_masdar.txt"

// ArabicVerbToMafoulMutlaqFilter ports org.languagetool.rules.ar.filters.ArabicVerbToMafoulMutlaqFilter.
//
// Accept path (Java king):
//  1. verb2masdarList = loadFromPath("/ar/arabic_verb_masdar.txt")
//  2. verb = args["verb"], adj = args["adj"]
//  3. tagger.getLemmas(patternTokens[0], "verb")
//  4. synthesizer.inflectAdjectiveTanwinNasb(adj, false/true)
//  5. for each lemma: verb2masdarList.get(lemma); for each msdr: inflectMafoulMutlq;
//     fem adj if msdr ends with TEH_MARBUTA else masc
//  6. suggestion "verb + ' ' + msdr + ' ' + adj"; List.contains dedupe
//
// Note: Java field verb2masdar (inline HashMap) is dead unused on the accept path — omitted.
type ArabicVerbToMafoulMutlaqFilter struct {
	// Verb2MasdarList ports verb2masdarList (file-backed SimpleReplaceDataLoader).
	Verb2MasdarList map[string][]string
	// tagger ports private final ArabicTagger tagger = new ArabicTagger().
	tagger *ar_tag.ArabicTagger
	// synthesizer ports private final ArabicSynthesizer synthesizer = new ArabicSynthesizer(new Arabic()).
	synthesizer *ar_synth.ArabicSynthesizer
}

// NewArabicVerbToMafoulMutlaqFilter ports the Java ctor: verb2masdarList = loadFromPath(FILE_NAME).
// Wires real ArabicTagger + ArabicSynthesizer (Java always constructs both).
func NewArabicVerbToMafoulMutlaqFilter() *ArabicVerbToMafoulMutlaqFilter {
	ar_tag.EnsureDefaultArabicTagger()
	return &ArabicVerbToMafoulMutlaqFilter{
		Verb2MasdarList: loadOfficialVerbMasdarMap(),
		tagger:          ar_tag.DefaultArabicTagger,
		synthesizer:     ar_synth.Instance(),
	}
}

// LoadVerbMasdarFromPath ports ArabicVerbToMafoulMutlaqFilter.loadFromPath (static).
// Prefer the ctor path (embedded official bytes); this loads via SimpleReplaceDataLoader
// from the rules classpath name for test twins.
func LoadVerbMasdarFromPath(path string) map[string][]string {
	if path == ArabicVerbMasdarFileName || path == "arabic_verb_masdar.txt" {
		return loadOfficialVerbMasdarMap()
	}
	return loadARSimpleReplaceMap(pathBase(path))
}

// AcceptRuleMatch ports ArabicVerbToMafoulMutlaqFilter.acceptRuleMatch.
// Args: verb, adj. Verb lemmas from tagger.getLemmas(patternTokens[0], "verb").
func (f *ArabicVerbToMafoulMutlaqFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	verb := arguments["verb"]
	adj := arguments["adj"]

	// Java: no bounds check (NPE if short). Go fail-closed empty match for nil/empty slice only.
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	if len(patternTokens) < 1 || patternTokens[0] == nil {
		return out
	}

	verbLemmas := f.tagger.GetLemmas(patternTokens[0], "verb")

	// generate multiple masdar from verb lemmas list
	var inflectedMasdarList []string
	var inflectedAdjList []string
	inflectedAdjMasculine := f.synthesizer.InflectAdjectiveTanwinNasb(adj, false)
	inflectedAdjFeminin := f.synthesizer.InflectAdjectiveTanwinNasb(adj, true)
	for _, lemma := range verbLemmas {
		// get suggested masdars lemmas — exact map get (Java loadWords keys only)
		msdrLemmaList := f.Verb2MasdarList[lemma]
		if msdrLemmaList == nil {
			continue
		}
		for _, msdr := range msdrLemmaList {
			// Java: if (msdr != null) — Go string is never null; empty still processes (data has none).
			inflectedMasdar := f.synthesizer.InflectMafoulMutlq(msdr)
			inflectedMasdarList = append(inflectedMasdarList, inflectedMasdar)
			inflectedAdj := inflectedAdjMasculine
			if strings.HasSuffix(msdr, string(tools.ArabicTehMarbuta)) {
				inflectedAdj = inflectedAdjFeminin
			}
			inflectedAdjList = append(inflectedAdjList, inflectedAdj)
		}
	}

	// Java: List.contains order-preserving dedupe
	var suggestionPhrases []string
	for i, msdr := range inflectedMasdarList {
		sugPhrase := verb + " " + msdr + " " + inflectedAdjList[i]
		if !containsString(suggestionPhrases, sugPhrase) {
			suggestionPhrases = append(suggestionPhrases, sugPhrase)
		}
	}
	if len(suggestionPhrases) > 0 {
		out.SetSuggestedReplacements(suggestionPhrases)
	}
	return out
}
