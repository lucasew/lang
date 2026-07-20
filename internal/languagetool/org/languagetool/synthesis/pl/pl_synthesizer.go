package pl

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java PolishSynthesizer constants.
const (
	potentialNegationTag = ":aff"
	negationTag          = ":neg"
	compTag              = "com"
	supTag               = "sup"
)

var (
	// segments matching .*[a-z]\.[a-z].* expand for setpos (getPosTagCorrection).
	plPosDotSegment = regexp.MustCompile(`.*[a-z]\.[a-z].*`)
	negationTagRE   = regexp.MustCompile(negationTag)
)

// PolishSynthesizer ports synthesis.pl.PolishSynthesizer.
type PolishSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewPolishSynthesizer(manual *synthesis.ManualSynthesizer) *PolishSynthesizer {
	base := synthesis.NewBaseSynthesizer("pl", manual)
	base.ResourceFileName = "/pl/polish_synth.dict"
	base.TagFileName = "/pl/polish_tags.txt"
	return &PolishSynthesizer{BaseSynthesizer: base}
}

// INSTANCE ports PolishSynthesizer.INSTANCE.
var INSTANCE = NewPolishSynthesizer(nil)

// Synthesize ports PolishSynthesizer.synthesize(token, posTag).
// null posTag → nil (Java null); + in tag forces regexp path.
func (s *PolishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if posTag == "" {
		// Java: if (posTag == null) return null; empty string is not null — treat as miss
		return nil, nil
	}
	if token == nil {
		return nil, nil
	}
	isNegated := isNegatedPL(token, posTag)
	if strings.Contains(posTag, "+") {
		return s.SynthesizeRE(token, posTag, true)
	}
	return s.getWordForms(token, posTag, isNegated), nil
}

// SynthesizeRE ports PolishSynthesizer.synthesize(token, pos, posTagRegExp).
func (s *PolishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, pos string, posTagRegExp bool) ([]string, error) {
	if pos == "" {
		return nil, nil
	}
	if token == nil {
		return nil, nil
	}
	if !posTagRegExp {
		return s.Synthesize(token, pos)
	}
	posTag := pos
	isNegated := isNegatedPL(token, posTag)
	if isNegated {
		posTag = strings.ReplaceAll(posTag, negationTag, potentialNegationTag+"?")
	}
	// Java: Pattern.compile(posTag.replace('+', '|'))
	pattern := strings.ReplaceAll(posTag, "+", "|")
	re, err := regexp.Compile("^(?:" + pattern + ")$")
	if err != nil {
		// Java: printStackTrace and return what was collected (empty)
		return []string{}, nil
	}
	seen := map[string]struct{}{}
	var results []string
	for _, tag := range s.possibleTags() {
		if !re.MatchString(tag) {
			continue
		}
		for _, f := range s.getWordForms(token, tag, isNegated) {
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			results = append(results, f)
		}
	}
	return results, nil
}

// isNegatedPL ports Java isNegated calculation (operator precedence && before ||):
//
//	posTag has :neg || (token POS has :neg && posTag lacks com && posTag lacks sup)
func isNegatedPL(token *languagetool.AnalyzedToken, posTag string) bool {
	if token == nil {
		return false
	}
	tokenPOS := ""
	if token.GetPOSTag() != nil {
		tokenPOS = *token.GetPOSTag()
	}
	// Java indexOf > 0 (not found is -1; at index 0 is false)
	a := strings.Index(posTag, negationTag) > 0
	b := strings.Index(tokenPOS, negationTag) > 0
	c := !(strings.Index(posTag, compTag) > 0)
	d := !(strings.Index(posTag, supTag) > 0)
	return a || (b && c && d)
}

// getWordForms ports PolishSynthesizer.getWordForms.
// Negated: lookup lemma|:aff (neg→aff) and prefix "nie".
func (s *PolishSynthesizer) getWordForms(token *languagetool.AnalyzedToken, posTag string, isNegated bool) []string {
	lemma := ""
	if token.GetLemma() != nil {
		lemma = *token.GetLemma()
	}
	if lemma == "" {
		lemma = token.GetToken()
	}
	if isNegated {
		affTag := negationTagRE.ReplaceAllString(posTag, potentialNegationTag)
		var forms []string
		for _, stem := range s.lookupLemmaTag(lemma, affTag) {
			forms = append(forms, "nie"+stem)
		}
		return forms
	}
	return s.lookupLemmaTag(lemma, posTag)
}

func (s *PolishSynthesizer) lookupLemmaTag(lemma, posTag string) []string {
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

func (s *PolishSynthesizer) possibleTags() []string {
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

// GetPosTagCorrection ports PolishSynthesizer.getPosTagCorrection.
func (s *PolishSynthesizer) GetPosTagCorrection(posTag string) string {
	if !strings.Contains(posTag, ".") {
		return posTag
	}
	tags := strings.Split(posTag, ":")
	pos := -1
	for i, t := range tags {
		if plPosDotSegment.MatchString(t) {
			// Java: Pattern.LITERAL "." → ".*|.*"
			tags[i] = "(.*" + strings.ReplaceAll(t, ".", ".*|.*") + ".*)"
			pos = i
		}
	}
	if pos == -1 {
		return posTag
	}
	return strings.Join(tags, ":")
}

var _ synthesis.Synthesizer = (*PolishSynthesizer)(nil)
