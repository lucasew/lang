package filters

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	ar_synth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	ar_tag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/ar"
)

// ArabicMasdarVerbFileName ports ArabicMasdarToVerbFilter.FILE_NAME.
const ArabicMasdarVerbFileName = "/ar/arabic_masdar_verb.txt"

// ArabicMasdarToVerbFilter ports org.languagetool.rules.ar.filters.ArabicMasdarToVerbFilter.
//
// Accept path (Java king):
//  1. masdar2verbList = loadFromPath("/ar/arabic_masdar_verb.txt")
//  2. tagger.getLemmas(patternTokens[0], "verb") → filterLemmas (authorizeLemma only "قَامَ")
//  3. tagger.getLemmas(patternTokens[1], "masdar")
//  4. for each aux reading with authorized lemma × masdar lemmas × map: synthesizer.inflectLemmaLike
//  5. HashSet dedupe → suggested replacements
//
// Note: Java field masdar2verb (inline HashMap) is dead unused on the accept path — omitted.
type ArabicMasdarToVerbFilter struct {
	// Masdar2VerbList ports masdar2verbList (file-backed SimpleReplaceDataLoader).
	Masdar2VerbList map[string][]string
	// tagger ports private final ArabicTagger tagger = new ArabicTagger().
	tagger *ar_tag.ArabicTagger
	// synthesizer ports private final ArabicSynthesizer synthesizer = new ArabicSynthesizer(new Arabic()).
	synthesizer *ar_synth.ArabicSynthesizer
	// authorizeLemma ports final List authorizeLemma = { "قَامَ" }.
	authorizeLemma []string
}

// NewArabicMasdarToVerbFilter ports the Java ctor: masdar2verbList = loadFromPath(FILE_NAME).
// Wires real ArabicTagger + ArabicSynthesizer (Java always constructs both).
func NewArabicMasdarToVerbFilter() *ArabicMasdarToVerbFilter {
	ar_tag.EnsureDefaultArabicTagger()
	return &ArabicMasdarToVerbFilter{
		Masdar2VerbList: loadOfficialMasdarVerbMap(),
		tagger:          ar_tag.DefaultArabicTagger,
		synthesizer:     ar_synth.Instance(),
		authorizeLemma:  []string{"قَامَ"},
	}
}

// LoadFromPath ports ArabicMasdarToVerbFilter.loadFromPath (static).
// Prefer the ctor path (embedded official bytes); this loads via SimpleReplaceDataLoader from a reader path name for test twins.
func LoadFromPath(path string) map[string][]string {
	// Java: new SimpleReplaceDataLoader().loadWords(path) from rules classpath.
	// Go twin of FILE_NAME only — same embed as ctor.
	if path == ArabicMasdarVerbFileName || path == "arabic_masdar_verb.txt" {
		return loadOfficialMasdarVerbMap()
	}
	// Other paths: discover under inspiration rules/ar (same layout as Java module resources).
	return loadARSimpleReplaceMap(pathBase(path))
}

func pathBase(path string) string {
	// strip leading /ar/ or directory so discoverARRulesFile can find basename
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}

// filterLemmas ports ArabicMasdarToVerbFilter.filterLemmas — keep only authorizeLemma entries present in lemmas.
func (f *ArabicMasdarToVerbFilter) filterLemmas(lemmas []string) []string {
	if f == nil {
		return nil
	}
	in := map[string]struct{}{}
	for _, l := range lemmas {
		in[l] = struct{}{}
	}
	var filtered []string
	for _, lem := range f.authorizeLemma {
		if _, ok := in[lem]; ok {
			filtered = append(filtered, lem)
		}
	}
	return filtered
}

// AcceptRuleMatch ports ArabicMasdarToVerbFilter.acceptRuleMatch.
// patternTokens[0]=auxiliary verb, [1]=masdar noun.
func (f *ArabicMasdarToVerbFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	// Java extracts then leaves unused for suggestion generation once tokens are tagged.
	_ = arguments["verb"] // الفعل قام أو ما شابهه
	_ = arguments["noun"] // masdar

	// Java: no bounds check (NPE if short). Go fail-closed empty match for nil slices only.
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	if len(patternTokens) < 2 || patternTokens[0] == nil || patternTokens[1] == nil {
		return out
	}

	// filter tokens which have a lemma; only auxiliary lemmas authorized
	auxVerbLemmasAll := f.tagger.GetLemmas(patternTokens[0], "verb")
	auxVerbLemmas := f.filterLemmas(auxVerbLemmasAll)
	masdarLemmas := f.tagger.GetLemmas(patternTokens[1], "masdar")

	// generate multiple verb from masdar lemmas list
	var verbList []string

	// if the auxiliary verb has many lemmas, filter authorized lemma only
	// the first token: auxiliary verb
	auxSet := map[string]struct{}{}
	for _, l := range auxVerbLemmas {
		auxSet[l] = struct{}{}
	}
	for _, auxVerbToken := range patternTokens[0].GetReadings() {
		if auxVerbToken == nil {
			continue
		}
		lem := ""
		if auxVerbToken.GetLemma() != nil {
			lem = *auxVerbToken.GetLemma()
		}
		// if the token has an authorized lemma
		if _, ok := auxSet[lem]; !ok {
			// Java: auxVerbLemmas.contains(auxVerbToken.getLemma()) — null lemma only if list has null
			continue
		}
		// for all masdar lemmas
		for _, lemma := range masdarLemmas {
			verbLemmaList := f.Masdar2VerbList[lemma] // exact get — no RemoveTashkeel invent
			if verbLemmaList == nil {
				continue
			}
			// if verb, inflect verd according to auxiliary verb inflection
			for _, vrbLem := range verbLemmaList {
				inflected := f.synthesizer.InflectLemmaLike(vrbLem, auxVerbToken)
				verbList = append(verbList, inflected...)
			}
		}
	}

	// remove duplicates (Java: new ArrayList<>(new HashSet<>(verbList)); order not guaranteed)
	seen := map[string]struct{}{}
	var deduped []string
	for _, verb := range verbList {
		if _, ok := seen[verb]; ok {
			continue
		}
		seen[verb] = struct{}{}
		deduped = append(deduped, verb)
	}
	// generate suggestion
	if len(deduped) > 0 {
		out.SetSuggestedReplacements(deduped)
	}
	return out
}
