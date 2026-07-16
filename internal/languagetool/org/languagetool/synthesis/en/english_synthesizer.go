package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

const (
	EnglishSynthResource = "/en/english_synth.dict"
	EnglishTagsFile      = "/en/english_tags.txt"
	EnglishSorFile       = "/en/en.sor"

	// Special synthesizer tags.
	AddDeterminer    = "+DT"
	AddIndDeterminer = "+INDT"
)

// EnglishSynthesizer ports org.languagetool.synthesis.en.EnglishSynthesizer.
type EnglishSynthesizer struct {
	*synthesis.BaseSynthesizer
	// AorAn chooses "a"/"an" for +INDT/+DT (pluggable).
	AorAn func(word string) string
}

func NewEnglishSynthesizer(manual *synthesis.ManualSynthesizer) *EnglishSynthesizer {
	base := synthesis.NewBaseSynthesizer("en", manual)
	base.ResourceFileName = EnglishSynthResource
	base.TagFileName = EnglishTagsFile
	return &EnglishSynthesizer{
		BaseSynthesizer: base,
		AorAn: func(word string) string {
			// simple vowel heuristic
			w := strings.ToLower(strings.TrimSpace(word))
			if w == "" {
				return "a"
			}
			switch w[0] {
			case 'a', 'e', 'i', 'o', 'u':
				return "an"
			default:
				return "a"
			}
		},
	}
}

// SynthesizeRE extends base with +DT / +INDT special tags.
func (s *EnglishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, posTagRegExp bool) ([]string, error) {
	if token == nil {
		return nil, nil
	}
	word := token.GetToken()
	if lemma := token.GetLemma(); lemma != nil && *lemma != "" {
		word = *lemma
	}
	switch posTag {
	case AddDeterminer:
		art := "a"
		if s.AorAn != nil {
			art = s.AorAn(word)
		}
		return []string{art + " " + word, "the " + word}, nil
	case AddIndDeterminer:
		art := "a"
		if s.AorAn != nil {
			art = s.AorAn(word)
		}
		return []string{art + " " + word}, nil
	default:
		return s.BaseSynthesizer.SynthesizeRE(token, posTag, posTagRegExp)
	}
}

func (s *EnglishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.SynthesizeRE(token, posTag, false)
}

var _ synthesis.Synthesizer = (*EnglishSynthesizer)(nil)
