package is

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const IcelandicTaggerDictPath = "/is/icelandic.dict"

// IcelandicTagger ports org.languagetool.tagging.is.IcelandicTagger.
type IcelandicTagger struct {
	*tagging.BaseTagger
}

func NewIcelandicTagger(wt tagging.WordTagger) *IcelandicTagger {
	return &IcelandicTagger{BaseTagger: tagging.NewBaseTagger(wt, IcelandicTaggerDictPath, "is", false)}
}
