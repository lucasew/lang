package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// WhiteSpaceBeforeParagraphEnd wraps the core WhiteSpaceBeforeParagraphEnd for this language.
type WhiteSpaceBeforeParagraphEnd struct {
	*rules.WhiteSpaceBeforeParagraphEnd
}

func NewWhiteSpaceBeforeParagraphEnd(messages map[string]string) *WhiteSpaceBeforeParagraphEnd {
	return &WhiteSpaceBeforeParagraphEnd{WhiteSpaceBeforeParagraphEnd: rules.NewWhiteSpaceBeforeParagraphEnd(messages)}
}

func (r *WhiteSpaceBeforeParagraphEnd) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.WhiteSpaceBeforeParagraphEnd.MatchList(sentences)
}
