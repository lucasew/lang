package ta

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const TamilTaggerDictPath = "/ta/tamil.dict"

// TamilTagger ports org.languagetool.tagging.ta.TamilTagger.
type TamilTagger struct {
	*tagging.BaseTagger
}

func NewTamilTagger(wt tagging.WordTagger) *TamilTagger {
	return &TamilTagger{BaseTagger: tagging.NewBaseTagger(wt, TamilTaggerDictPath, "ta", false)}
}
