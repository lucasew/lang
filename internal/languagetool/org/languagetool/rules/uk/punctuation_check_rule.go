package uk

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PunctuationCheckRule ports org.languagetool.rules.uk.PunctuationCheckRule.
//
// Java: extends AbstractPunctuationCheckRule; overrides isPunctsJoinOk / isPunctuation.
// Id, description, and category come from AbstractPunctuationCheckRule
// (PUNCTUATION_GENERIC_CHECK, Categories.PUNCTUATION, fixed English description).
type PunctuationCheckRule struct {
	*rules.AbstractPunctuationCheckRule
}

// Java static finals — Pattern.matches is full-string match (Go: MatchString with ^$).
// PATTERN_3 already has anchors in Java; PATTERN_1/2 rely on matches().
var (
	// PATTERN_1 = Pattern.compile("([,:] | *- |,- | ) *");
	ukPunctuationCheckPattern1 = regexp.MustCompile(`^([,:] | *- |,- | ) *$`)
	// PATTERN_2 = Pattern.compile("([.!?]|!!!|\\?\\?\\?|\\?!!|!\\.\\.|\\?\\.\\.|\\.\\.\\.) *");
	ukPunctuationCheckPattern2 = regexp.MustCompile(`^([.!?]|!!!|\?\?\?|\?!!|!\.\.|\?\.\.|\.\.\.) *$`)
	// PATTERN_3 = Pattern.compile("^[.,!?: -]$");
	ukPunctuationCheckPattern3 = regexp.MustCompile(`^[.,!?: -]$`)
)

// NewPunctuationCheckRule ports PunctuationCheckRule(ResourceBundle messages).
// Java: super(messages); // setCategory commented out in UK subclass (abstract already sets PUNCTUATION).
func NewPunctuationCheckRule(messages map[string]string) *PunctuationCheckRule {
	base := &rules.AbstractPunctuationCheckRule{
		Messages: messages,
	}
	// Java AbstractPunctuationCheckRule ctor: Categories.PUNCTUATION + getDescription string.
	rules.InitPunctuationCheckMeta(base, messages)
	r := &PunctuationCheckRule{AbstractPunctuationCheckRule: base}
	// Wire abstract hooks to UK overrides (Java protected abstract methods).
	base.IsPunctsJoinOk = r.isPunctsJoinOk
	base.IsPunctuation = r.isPunctuation
	return r
}

// isPunctsJoinOk ports PunctuationCheckRule.isPunctsJoinOk (final in Java).
// return PATTERN_1.matcher(tokens).matches() || PATTERN_2.matcher(tokens).matches();
func (r *PunctuationCheckRule) isPunctsJoinOk(tokens string) bool {
	// we ignore duplicated spaces - too many errors (Java comment)
	return ukPunctuationCheckPattern1.MatchString(tokens) || ukPunctuationCheckPattern2.MatchString(tokens)
}

// isPunctuation ports PunctuationCheckRule.isPunctuation (final in Java).
// return PATTERN_3.matcher(token).matches();
func (r *PunctuationCheckRule) isPunctuation(token string) bool {
	return ukPunctuationCheckPattern3.MatchString(token)
}

// Match ports AbstractPunctuationCheckRule.match via the UK hooks.
func (r *PunctuationCheckRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.AbstractPunctuationCheckRule.Match(sentence)
}
