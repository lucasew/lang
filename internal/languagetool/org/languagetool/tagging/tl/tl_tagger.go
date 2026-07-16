package tl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const TagalogTaggerDictPath = "/tl/tagalog.dict"

// TagalogTagger ports org.languagetool.tagging.tl.TagalogTagger.
type TagalogTagger struct {
	*tagging.BaseTagger
}

func NewTagalogTagger(wt tagging.WordTagger) *TagalogTagger {
	return &TagalogTagger{BaseTagger: tagging.NewBaseTagger(wt, TagalogTaggerDictPath, "tl", false)}
}
