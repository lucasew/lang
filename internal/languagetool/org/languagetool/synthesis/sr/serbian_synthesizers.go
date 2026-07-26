package sr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// Java SerbianSynthesizer / EkavianSynthesizer / JekavianSynthesizer resource paths.
// Java SerbianSynthesizer:
//   DICTIONARY_PATH = "/sr/dictionary/"
//   TAGS_FILE_NAME  = DICTIONARY_PATH + "serbian_synth_tags.txt"
// Java EkavianSynthesizer RESOURCE_FILENAME = DICTIONARY_PATH + "ekavian/serbian_synth.dict"
// Java JekavianSynthesizer RESOURCE_FILENAME = DICTIONARY_PATH + "jekavian/serbian_synth.dict"
//
// Java SerbianSynthesizer calls super(resource, tags) 2-arg; current BaseSynthesizer
// needs langShortCode — port intent is super(resource, tags, "sr").
const (
	DictionaryPath    = "/sr/dictionary/"
	TagsFileName      = DictionaryPath + "serbian_synth_tags.txt"
	EkavianSynthDict  = DictionaryPath + "ekavian/serbian_synth.dict"
	JekavianSynthDict = DictionaryPath + "jekavian/serbian_synth.dict"
)

// SerbianSynthesizer ports org.languagetool.synthesis.sr.SerbianSynthesizer.
// Thin BaseSynthesizer subclass (no method overrides in Java).
type SerbianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

// NewSerbianSynthesizer ports SerbianSynthesizer(resourceFileName, tagFileName)
// with langShortCode "sr" (see BaseSynthesizer 3-arg intent above).
func NewSerbianSynthesizer(resourceFileName, tagFileName string, manual *synthesis.ManualSynthesizer) *SerbianSynthesizer {
	base := synthesis.NewBaseSynthesizer("sr", manual)
	base.ResourceFileName = resourceFileName
	base.TagFileName = tagFileName
	return &SerbianSynthesizer{BaseSynthesizer: base}
}

// Synthesize implements synthesis.Synthesizer (Java inherits BaseSynthesizer.synthesize).
func (s *SerbianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer (Java inherits BaseSynthesizer RE path).
func (s *SerbianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.BaseSynthesizer == nil {
		return nil, nil
	}
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*SerbianSynthesizer)(nil)

// EkavianSynthesizer ports org.languagetool.synthesis.sr.EkavianSynthesizer.
// No method overrides — only resource path constants.
type EkavianSynthesizer struct{ *SerbianSynthesizer }

// NewEkavianSynthesizer ports EkavianSynthesizer() path wiring (no official dict load).
// Java: super(RESOURCE_FILENAME, TAGS_FILE_NAME) → super(..., "sr").
func NewEkavianSynthesizer(manual *synthesis.ManualSynthesizer) *EkavianSynthesizer {
	return &EkavianSynthesizer{SerbianSynthesizer: NewSerbianSynthesizer(EkavianSynthDict, TagsFileName, manual)}
}

// Synthesize implements synthesis.Synthesizer.
func (s *EkavianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.SerbianSynthesizer == nil {
		return nil, nil
	}
	return s.SerbianSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer.
func (s *EkavianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.SerbianSynthesizer == nil {
		return nil, nil
	}
	return s.SerbianSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*EkavianSynthesizer)(nil)

// JekavianSynthesizer ports org.languagetool.synthesis.sr.JekavianSynthesizer.
// No method overrides — only resource path constants (jekavian/serbian_synth.dict).
type JekavianSynthesizer struct{ *SerbianSynthesizer }

// NewJekavianSynthesizer ports JekavianSynthesizer() path wiring.
func NewJekavianSynthesizer(manual *synthesis.ManualSynthesizer) *JekavianSynthesizer {
	return &JekavianSynthesizer{SerbianSynthesizer: NewSerbianSynthesizer(JekavianSynthDict, TagsFileName, manual)}
}

// Synthesize implements synthesis.Synthesizer.
func (s *JekavianSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	if s == nil || s.SerbianSynthesizer == nil {
		return nil, nil
	}
	return s.SerbianSynthesizer.Synthesize(token, posTag)
}

// SynthesizeRE implements synthesis.Synthesizer.
func (s *JekavianSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	if s == nil || s.SerbianSynthesizer == nil {
		return nil, nil
	}
	return s.SerbianSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*JekavianSynthesizer)(nil)
