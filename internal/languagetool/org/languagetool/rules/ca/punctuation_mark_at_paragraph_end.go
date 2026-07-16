package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PunctuationMarkAtParagraphEnd wraps the core PunctuationMarkAtParagraphEnd for this language.
type PunctuationMarkAtParagraphEnd struct {
	*rules.PunctuationMarkAtParagraphEnd
}

func NewPunctuationMarkAtParagraphEnd(messages map[string]string) *PunctuationMarkAtParagraphEnd {
	return &PunctuationMarkAtParagraphEnd{PunctuationMarkAtParagraphEnd: rules.NewPunctuationMarkAtParagraphEnd(messages)}
}

func (r *PunctuationMarkAtParagraphEnd) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.PunctuationMarkAtParagraphEnd.MatchList(sentences)
}
