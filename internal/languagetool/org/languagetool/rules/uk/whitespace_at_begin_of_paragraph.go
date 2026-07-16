package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WhiteSpaceAtBeginOfParagraph wraps the core rule for this language.
type WhiteSpaceAtBeginOfParagraph struct {
	*rules.WhiteSpaceAtBeginOfParagraph
}

func NewWhiteSpaceAtBeginOfParagraph(messages map[string]string) *WhiteSpaceAtBeginOfParagraph {
	return &WhiteSpaceAtBeginOfParagraph{WhiteSpaceAtBeginOfParagraph: rules.NewWhiteSpaceAtBeginOfParagraph(messages)}
}

func (r *WhiteSpaceAtBeginOfParagraph) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WhiteSpaceAtBeginOfParagraph.Match(sentence)
}
