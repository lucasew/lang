package sr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"

const (
	EkavianSynthDict   = "/sr/dictionary/ekavian/serbian_synth.dict"
	JekavianSynthDict  = "/sr/dictionary/jekavian/serbian_synth.dict"
	SerbianTagsFile    = "/sr/serbian_tags.txt"
)

// SerbianSynthesizer ports org.languagetool.synthesis.sr.SerbianSynthesizer (Ekavian default).
type SerbianSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewSerbianSynthesizer(manual *synthesis.ManualSynthesizer) *SerbianSynthesizer {
	return NewSerbianSynthesizerWith(manual, EkavianSynthDict)
}

func NewSerbianSynthesizerWith(manual *synthesis.ManualSynthesizer, dict string) *SerbianSynthesizer {
	base := synthesis.NewBaseSynthesizer("sr", manual)
	base.ResourceFileName = dict
	base.TagFileName = SerbianTagsFile
	return &SerbianSynthesizer{BaseSynthesizer: base}
}

// EkavianSynthesizer ports org.languagetool.synthesis.sr.EkavianSynthesizer.
type EkavianSynthesizer struct{ *SerbianSynthesizer }

func NewEkavianSynthesizer(manual *synthesis.ManualSynthesizer) *EkavianSynthesizer {
	return &EkavianSynthesizer{SerbianSynthesizer: NewSerbianSynthesizerWith(manual, EkavianSynthDict)}
}

// JekavianSynthesizer ports org.languagetool.synthesis.sr.JekavianSynthesizer.
type JekavianSynthesizer struct{ *SerbianSynthesizer }

func NewJekavianSynthesizer(manual *synthesis.ManualSynthesizer) *JekavianSynthesizer {
	return &JekavianSynthesizer{SerbianSynthesizer: NewSerbianSynthesizerWith(manual, JekavianSynthDict)}
}
