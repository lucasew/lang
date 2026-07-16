package translation

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Inflector ports org.languagetool.rules.en.translation.Inflector.
// Inflects English words according to a German POS tag via a synthesizer.
type Inflector struct {
	Synth synthesis.Synthesizer
}

func NewInflector(synth synthesis.Synthesizer) *Inflector {
	return &Inflector{Synth: synth}
}

// Inflect applies dePosTag-driven inflection to the last word of enToken.
func (inf *Inflector) Inflect(enToken, dePosTag string) []string {
	parts := strings.Fields(strings.Replace(enToken, "to ", "", 1))
	if len(parts) == 0 {
		return nil
	}
	lastForms := inf.inflectSingleWord(parts[len(parts)-1], dePosTag)
	start := ""
	if len(parts) > 1 {
		start = strings.Join(parts[:len(parts)-1], " ")
	}
	var result []string
	for _, last := range lastForms {
		if start != "" {
			result = append(result, strings.TrimSpace(start+" "+last))
		} else {
			result = append(result, last)
		}
	}
	return result
}

func (inf *Inflector) inflectSingleWord(enToken, dePosTag string) []string {
	if dePosTag == "" {
		return []string{enToken}
	}
	switch {
	case matchRE(dePosTag, `SUB.*PLU.*`):
		return inf.getForms(enToken, "NNP?S")
	case matchRE(dePosTag, `VER:3:SIN:PRÄ.*`):
		return inf.getForms(enToken, "VBZ")
	case matchRE(dePosTag, `VER:3:SIN:PRT:.*`):
		return inf.getForms(enToken, "VBD")
	case matchRE(dePosTag, `PA1:PRD:GRU:VER`):
		return inf.getForms(enToken, "VBG")
	case matchRE(dePosTag, `PA2:PRD:GRU:VER`):
		return inf.getForms(enToken, "VBN")
	case matchRE(dePosTag, `ADJ:PRD:KOM|ADJ:.*:KOM.*`):
		return inf.getForms(enToken, "JJR")
	case matchRE(dePosTag, `ADJ:.*:SUP.*`):
		return inf.getForms(enToken, "JJS")
	default:
		return []string{enToken}
	}
}

func (inf *Inflector) getForms(enToken, posTagRegex string) []string {
	if inf == nil || inf.Synth == nil {
		return []string{enToken}
	}
	tok := languagetool.NewAnalyzedToken(enToken, strPtr("fake-value"), strPtr(enToken))
	forms, err := inf.Synth.SynthesizeRE(tok, posTagRegex, true)
	if err != nil || len(forms) == 0 {
		return []string{enToken}
	}
	return forms
}

func matchRE(s, re string) bool {
	ok, err := regexp.MatchString(re, s)
	return err == nil && ok
}

func strPtr(s string) *string { return &s }
