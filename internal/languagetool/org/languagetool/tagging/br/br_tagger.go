package br

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const BretonTaggerDictPath = "/br/breton.dict"

// BretonTagger ports org.languagetool.tagging.br.BretonTagger.
type BretonTagger struct {
	*tagging.BaseTagger
}

func NewBretonTagger(wt tagging.WordTagger) *BretonTagger {
	return &BretonTagger{BaseTagger: tagging.NewBaseTagger(wt, BretonTaggerDictPath, "br", false)}
}
