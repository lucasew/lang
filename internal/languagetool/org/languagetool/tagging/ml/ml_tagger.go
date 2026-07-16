package ml

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const MalayalamTaggerDictPath = "/ml/malayalam.dict"

// MalayalamTagger ports org.languagetool.tagging.ml.MalayalamTagger.
type MalayalamTagger struct {
	*tagging.BaseTagger
}

func NewMalayalamTagger(wt tagging.WordTagger) *MalayalamTagger {
	return &MalayalamTagger{BaseTagger: tagging.NewBaseTagger(wt, MalayalamTaggerDictPath, "ml", false)}
}
