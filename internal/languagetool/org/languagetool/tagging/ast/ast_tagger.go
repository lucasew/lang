package ast

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"

const AsturianTaggerDictPath = "/ast/asturian.dict"

// AsturianTagger ports org.languagetool.tagging.ast.AsturianTagger.
type AsturianTagger struct {
	*tagging.BaseTagger
}

func NewAsturianTagger(wt tagging.WordTagger) *AsturianTagger {
	return &AsturianTagger{BaseTagger: tagging.NewBaseTagger(wt, AsturianTaggerDictPath, "ast", false)}
}
