package sr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const (
	SerbianBaseDictPath    = "/sr/dictionary"
	EkavianDictionaryPath  = SerbianBaseDictPath + "/ekavian/serbian.dict"
	JekavianDictionaryPath = SerbianBaseDictPath + "/jekavian/serbian.dict"
)

// SerbianTagger ports org.languagetool.tagging.sr.SerbianTagger (Ekavian default).
type SerbianTagger struct {
	*tagging.BaseTagger
}

func NewSerbianTagger(wt tagging.WordTagger) *SerbianTagger {
	return NewSerbianTaggerWithPath(wt, EkavianDictionaryPath)
}

func NewSerbianTaggerWithPath(wt tagging.WordTagger, path string) *SerbianTagger {
	return &SerbianTagger{BaseTagger: tagging.NewBaseTagger(wt, path, "sr", false)}
}

// EkavianTagger ports org.languagetool.tagging.sr.EkavianTagger.
type EkavianTagger struct{ *SerbianTagger }

func NewEkavianTagger(wt tagging.WordTagger) *EkavianTagger {
	return &EkavianTagger{SerbianTagger: NewSerbianTaggerWithPath(wt, EkavianDictionaryPath)}
}

// JekavianTagger ports org.languagetool.tagging.sr.JekavianTagger.
type JekavianTagger struct{ *SerbianTagger }

func NewJekavianTagger(wt tagging.WordTagger) *JekavianTagger {
	return &JekavianTagger{SerbianTagger: NewSerbianTaggerWithPath(wt, JekavianDictionaryPath)}
}
