package es

import (
	"regexp"
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java SpanishSynthesizer.RESOURCE_FILENAME / TAGS_FILE_NAME.
const (
	SpanishSynthDict = "/es/es-ES_synth.dict"
	SpanishTagsFile  = "/es/es-ES_tags.txt"
	SpanishSorFile   = "/es/es.sor"
)

// pLemmaSpace ports SpanishSynthesizer.pLemmaSpace: "([^ ]+) (.+)"
var pLemmaSpace = regexp.MustCompile(`^([^ ]+) (.+)$`)

// SpanishSynthesizer ports synthesis.es.SpanishSynthesizer.
type SpanishSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewSpanishSynthesizer(manual *synthesis.ManualSynthesizer) *SpanishSynthesizer {
	base := synthesis.NewBaseSynthesizer("es", manual)
	// Java: super("/es/es.sor", "/es/es-ES_synth.dict", "/es/es-ES_tags.txt", "es")
	base.SorFileName = SpanishSorFile
	base.ResourceFileName = SpanishSynthDict
	base.TagFileName = SpanishTagsFile
	return &SpanishSynthesizer{BaseSynthesizer: base}
}

// INSTANCE matches SpanishSynthesizer.INSTANCE (manual/lookup set by openers).
var INSTANCE = NewSpanishSynthesizer(nil)

// Synthesize ports SpanishSynthesizer.synthesize(token, posTag):
// spell-number → super; verb lemma "verb rest" → lookup verb + append " rest".
func (s *SpanishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if strings.HasPrefix(posTag, synthesis.SpellNumberTag) {
		return s.BaseSynthesizer.Synthesize(token, posTag)
	}
	lemma, toAdd := splitVerbLemma(token, posTag)
	results := s.lookupLemmaTag(lemma, posTag)
	return addWordsAfter(results, toAdd), nil
}

// SynthesizeRE ports SpanishSynthesizer.synthesize(token, posTag, posTagRegExp).
func (s *SpanishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if strings.HasPrefix(posTag, synthesis.SpellNumberTag) {
		return s.Synthesize(token, posTag)
	}
	if !posTagRegExp {
		return s.Synthesize(token, posTag)
	}
	lemma, toAdd := splitVerbLemma(token, posTag)
	re, err := regexp.Compile("^(?:" + posTag + ")$")
	if err != nil {
		// Java: log warning, return null
		return nil, nil
	}
	var results []string
	for _, tag := range s.possibleTags() {
		if re.MatchString(tag) {
			results = append(results, s.lookupLemmaTag(lemma, tag)...)
		}
	}
	return addWordsAfter(results, toAdd), nil
}

func splitVerbLemma(token *languagetool.AnalyzedToken, posTag string) (lemma, toAddAfter string) {
	lemma = ""
	if token != nil {
		if token.GetLemma() != nil {
			lemma = *token.GetLemma()
		}
		if lemma == "" {
			lemma = token.GetToken()
		}
	}
	toAddAfter = ""
	// Java: only when posTag.startsWith("V")
	if strings.HasPrefix(posTag, "V") {
		if m := pLemmaSpace.FindStringSubmatch(lemma); len(m) == 3 {
			lemma = m[1]
			toAddAfter = m[2]
		}
	}
	return lemma, toAddAfter
}

func addWordsAfter(results []string, toAddAfter string) []string {
	if toAddAfter == "" {
		return results
	}
	out := make([]string, 0, len(results))
	for _, r := range results {
		out = append(out, r+" "+toAddAfter)
	}
	return out
}

func (s *SpanishSynthesizer) lookupLemmaTag(lemma, posTag string) []string {
	if s == nil || s.BaseSynthesizer == nil || lemma == "" {
		return nil
	}
	var out []string
	if s.Lookup != nil {
		out = append(out, s.Lookup(lemma, posTag)...)
	}
	if s.Manual != nil {
		out = append(out, s.Manual.Lookup(lemma, posTag)...)
	}
	if s.Removal != nil {
		filtered := out[:0]
		for _, f := range out {
			removed := false
			for _, r := range s.Removal.Lookup(lemma, posTag) {
				if r == f {
					removed = true
					break
				}
			}
			if !removed {
				filtered = append(filtered, f)
			}
		}
		out = filtered
	}
	return out
}

func (s *SpanishSynthesizer) possibleTags() []string {
	if s == nil || s.BaseSynthesizer == nil {
		return nil
	}
	if len(s.PossibleTags) > 0 {
		return s.PossibleTags
	}
	if s.Manual != nil {
		var tags []string
		for t := range s.Manual.GetPossibleTags() {
			tags = append(tags, t)
		}
		return tags
	}
	return nil
}

// GetTargetPosTag ports SpanishSynthesizer.getTargetPosTag:
// sort with PostagComparator (Indicative > Imperative), return last.
func (s *SpanishSynthesizer) GetTargetPosTag(posTags []string, targetPosTag string) string {
	if len(posTags) == 0 {
		return targetPosTag
	}
	cp := append([]string(nil), posTags...)
	sort.SliceStable(cp, func(i, j int) bool {
		// Comparator.compare(a,b) < 0 means a before b.
		return postagCompare(cp[i], cp[j]) < 0
	})
	return cp[len(cp)-1]
}

// postagCompare ports SpanishSynthesizer.PostagComparator.
// VMIP3S0 before VMM02S0 (compare returns +150 so arg0>arg1 when sorting? Java Comparator:
// compare(a,b)>0 means a after b in ascending sort. compare(VMIP3S0, VMM02S0)=150 → VMIP after VMM.
// compare(VMM02S0, VMIP3S0)=-150 → VMM before VMIP. Sorted ascending: VMM then VMIP; last = VMIP.
func postagCompare(arg0, arg1 string) int {
	if len(arg0) > 4 && len(arg1) > 4 {
		if arg0 == "VMIP3S0" && arg1 == "VMM02S0" {
			return 150
		}
		if arg0 == "VMM02S0" && arg1 == "VMIP3S0" {
			return -150
		}
	}
	return 0
}

var _ synthesis.Synthesizer = (*SpanishSynthesizer)(nil)
