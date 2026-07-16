package lt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const LithuanianTaggerDictPath = "/lt/lithuanian.dict"

// LithuanianTagger ports org.languagetool.tagging.lt.LithuanianTagger.
type LithuanianTagger struct {
	*tagging.BaseTagger
}

func NewLithuanianTagger(wt tagging.WordTagger) *LithuanianTagger {
	return &LithuanianTagger{BaseTagger: tagging.NewBaseTagger(wt, LithuanianTaggerDictPath, "lt", false)}
}
