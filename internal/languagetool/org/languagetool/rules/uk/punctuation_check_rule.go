package uk

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PunctuationCheckRule ports org.languagetool.rules.uk.PunctuationCheckRule.
type PunctuationCheckRule struct {
	*rules.AbstractPunctuationCheckRule
}

var (
	ukPunctJoin1 = regexp.MustCompile(`^([,:] | *- |,- | ) *$`)
	ukPunctJoin2 = regexp.MustCompile(`^([.!?]|!!!|\?\?\?|\?!!|!\.\.|\?\.\.|\.\.\.) *$`)
	ukPunctToken = regexp.MustCompile(`^[.,!?: -]$`)
)

func NewPunctuationCheckRule(messages map[string]string) *PunctuationCheckRule {
	base := &rules.AbstractPunctuationCheckRule{
		Messages: messages,
		ID:       "PUNCTUATION_GENERIC_CHECK",
		IsPunctsJoinOk: func(tokens string) bool {
			return ukPunctJoin1.MatchString(tokens) || ukPunctJoin2.MatchString(tokens)
		},
		IsPunctuation: func(token string) bool {
			return ukPunctToken.MatchString(token)
		},
	}
	// Java AbstractPunctuationCheckRule: Categories.PUNCTUATION + English description
	rules.InitPunctuationCheckMeta(base, messages)
	return &PunctuationCheckRule{AbstractPunctuationCheckRule: base}
}

func (r *PunctuationCheckRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractPunctuationCheckRule.Match(sentence)
}
