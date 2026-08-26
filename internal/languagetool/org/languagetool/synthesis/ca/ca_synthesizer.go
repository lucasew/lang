package ca

import (
	"regexp"
	"sort"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java CatalanSynthesizer.verbTags by Language.getShortCodeWithCountryAndVariant().
var verbTags = map[string]string{
	"ca-ES":          "[0CXY12]",
	"ca-ES-valencia": "[0VXZ13567]",
	"ca-ES-balear":   "[0BYZ1247]",
}

// LemmasToIgnore ports CatalanSynthesizer.LemmasToIgnore (empty synthesis).
var LemmasToIgnore = map[string]struct{}{
	"enterar":   {},
	"sentar":    {},
	"conseguir": {},
	"alcançar":  {},
}

// pVerb ports CatalanSynthesizer.pVerb: V.*[CVBXYZ0123456]
var pVerb = regexp.MustCompile(`^V.*[CVBXYZ0123456]$`)

// pLemmaSpace ports "([^ ]+) (.+)"
var pLemmaSpace = regexp.MustCompile(`^([^ ]+) (.+)$`)

// CatalanSynthesizer ports synthesis.ca.CatalanSynthesizer.
type CatalanSynthesizer struct {
	*synthesis.BaseSynthesizer
	// LanguageCode is Java Language short code with country/variant (ca-ES, ca-ES-valencia, …).
	LanguageCode string
}

func NewCatalanSynthesizer(manual *synthesis.ManualSynthesizer) *CatalanSynthesizer {
	return NewCatalanSynthesizerForLang(manual, "ca-ES")
}

// NewCatalanSynthesizerForLang ports CatalanSynthesizer(Language) constructors
// (INSTANCE_CAT / INSTANCE_VAL / INSTANCE_BAL).
func NewCatalanSynthesizerForLang(manual *synthesis.ManualSynthesizer, langCode string) *CatalanSynthesizer {
	if langCode == "" {
		langCode = "ca-ES"
	}
	base := synthesis.NewBaseSynthesizer("ca", manual)
	base.ResourceFileName = "/ca/ca-ES_synth.dict"
	base.TagFileName = "/ca/ca-ES_tags.txt"
	base.SorFileName = "/ca/ca.sor"
	return &CatalanSynthesizer{BaseSynthesizer: base, LanguageCode: langCode}
}

// INSTANCE_CAT / INSTANCE_VAL / INSTANCE_BAL port Java static instances.
var (
	INSTANCE_CAT = NewCatalanSynthesizerForLang(nil, "ca-ES")
	INSTANCE_VAL = NewCatalanSynthesizerForLang(nil, "ca-ES-valencia")
	INSTANCE_BAL = NewCatalanSynthesizerForLang(nil, "ca-ES-balear")
)

// Synthesize ports CatalanSynthesizer.synthesize(token, posTag):
// always regex-matches posTag against possibleTags (Java Pattern.compile(posTag)).
func (s *CatalanSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if strings.HasPrefix(posTag, synthesis.SpellNumberTag) {
		return s.BaseSynthesizer.Synthesize(token, posTag)
	}
	lemma, toAdd := splitVerbLemma(token, posTag)
	results := s.matchLookup(lemma, posTag)
	// if not found, try verbs from a regional variant
	if len(results) == 0 && strings.HasPrefix(posTag, "V") {
		region := s.verbTagSuffix()
		if region != "" && len(posTag) > 0 {
			alt := posTag[:len(posTag)-1] + region
			return s.SynthesizeRE(token, alt, true)
		}
	}
	return addWordsAfter(results, toAdd), nil
}

// SynthesizeRE ports CatalanSynthesizer.synthesize(token, posTag, posTagRegExp).
func (s *CatalanSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if strings.HasPrefix(posTag, synthesis.SpellNumberTag) {
		return s.Synthesize(token, posTag)
	}
	if !posTagRegExp {
		return s.Synthesize(token, posTag)
	}
	lemma := lemmaOf(token)
	if _, skip := LemmasToIgnore[lemma]; skip {
		return []string{}, nil
	}
	toAdd := ""
	if strings.HasPrefix(posTag, "V") {
		if m := pLemmaSpace.FindStringSubmatch(lemma); len(m) == 3 {
			lemma = m[1]
			toAdd = m[2]
		}
	}
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
	// if not found, try verbs from the active regional variant
	if len(results) == 0 && pVerb.MatchString(posTag) {
		region := s.verbTagSuffix()
		if region != "" && len(posTag) > 0 {
			alt := posTag[:len(posTag)-1] + region
			re2, err2 := regexp.Compile("^(?:" + alt + ")$")
			if err2 == nil {
				for _, tag := range s.possibleTags() {
					if re2.MatchString(tag) {
						results = append(results, s.lookupLemmaTag(lemma, tag)...)
					}
				}
			}
		}
	}
	return addWordsAfter(results, toAdd), nil
}

func (s *CatalanSynthesizer) verbTagSuffix() string {
	code := "ca-ES"
	if s != nil && s.LanguageCode != "" {
		code = s.LanguageCode
	}
	return verbTags[code]
}

func (s *CatalanSynthesizer) matchLookup(lemma, posTagPattern string) []string {
	re, err := regexp.Compile("^(?:" + posTagPattern + ")$")
	if err != nil {
		return nil
	}
	var results []string
	for _, tag := range s.possibleTags() {
		if re.MatchString(tag) {
			results = append(results, s.lookupLemmaTag(lemma, tag)...)
		}
	}
	return results
}

func lemmaOf(token *languagetool.AnalyzedToken) string {
	if token == nil {
		return ""
	}
	if token.GetLemma() != nil && *token.GetLemma() != "" {
		return *token.GetLemma()
	}
	return token.GetToken()
}

func splitVerbLemma(token *languagetool.AnalyzedToken, posTag string) (lemma, toAddAfter string) {
	lemma = lemmaOf(token)
	toAddAfter = ""
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

func (s *CatalanSynthesizer) lookupLemmaTag(lemma, posTag string) []string {
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

func (s *CatalanSynthesizer) possibleTags() []string {
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

// GetPossibleTags ports CatalanSynthesizer.getPossibleTags.
func (s *CatalanSynthesizer) GetPossibleTags() []string {
	return s.possibleTags()
}

// GetTargetPosTag ports CatalanSynthesizer.getTargetPosTag (PostagComparator; last after sort).
func (s *CatalanSynthesizer) GetTargetPosTag(posTags []string, targetPosTag string) string {
	if len(posTags) == 0 {
		return targetPosTag
	}
	cp := append([]string(nil), posTags...)
	sort.SliceStable(cp, func(i, j int) bool {
		return caPostagCompare(cp[i], cp[j]) < 0
	})
	return cp[len(cp)-1]
}

// caPostagCompare ports CatalanSynthesizer.PostagComparator:
// priority 3 person > 1 person, Indicative > Subjunctive, special VMIP2P00/VMIS3S00.
func caPostagCompare(arg0, arg1 string) int {
	if len(arg0) > 4 && len(arg1) > 4 {
		if strings.Contains(arg0, "3S") && arg1 == "1S" {
			return 150
		}
		if strings.Contains(arg0, "1S") && strings.Contains(arg1, "3S") {
			return -150
		}
		if arg0 == "VMIP2P00" && arg1 == "VMIS3S00" {
			return 150
		}
		if arg1 == "VMIP2P00" && arg0 == "VMIS3S00" {
			return -150
		}
		if arg0[2] == 'I' && arg1[2] != 'I' {
			return 100
		}
		if arg1[2] == 'I' && arg0[2] != 'I' {
			return -100
		}
		if arg0[4] == '3' && arg1[4] == '1' {
			return 50
		}
		if arg1[4] == '1' && arg0[4] == '3' {
			return -50
		}
	}
	return 0
}

var _ synthesis.Synthesizer = (*CatalanSynthesizer)(nil)
