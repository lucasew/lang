package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PunctuationMarkAtParagraphEnd2 wraps the core rule for this language.
type PunctuationMarkAtParagraphEnd2 struct {
	*rules.PunctuationMarkAtParagraphEnd2
}

func NewPunctuationMarkAtParagraphEnd2(messages map[string]string) *PunctuationMarkAtParagraphEnd2 {
	return &PunctuationMarkAtParagraphEnd2{PunctuationMarkAtParagraphEnd2: rules.NewPunctuationMarkAtParagraphEnd2(messages)}
}

func (r *PunctuationMarkAtParagraphEnd2) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.PunctuationMarkAtParagraphEnd2.MatchList(sentences)
}
