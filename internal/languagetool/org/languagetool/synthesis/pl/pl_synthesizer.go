package pl

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java PolishSynthesizer: segments with a.z-style dots expand for setpos synthesis.
var (
	plPosDotSegment = regexp.MustCompile(`.*[a-z]\.[a-z].*`)
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

func (s *PolishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *PolishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

// GetPosTagCorrection ports PolishSynthesizer.getPosTagCorrection.
// Colon-separated tags: segments matching .*[a-z]\.[a-z].* become
// (.*a.*|.*z.*)-style alternatives (literal '.' → .*|.*).
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
