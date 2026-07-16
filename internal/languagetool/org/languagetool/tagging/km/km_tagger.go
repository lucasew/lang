package km

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const KhmerTaggerDictPath = "/km/khmer.dict"

// KhmerTagger ports org.languagetool.tagging.km.KhmerTagger.
type KhmerTagger struct {
	*tagging.BaseTagger
}

func NewKhmerTagger(wt tagging.WordTagger) *KhmerTagger {
	return &KhmerTagger{BaseTagger: tagging.NewBaseTagger(wt, KhmerTaggerDictPath, "km", false)}
}
