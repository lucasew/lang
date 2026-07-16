package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PatternRule ports a metadata surface of org.languagetool.rules.patterns.PatternRule
// (full matcher engine not yet wired).
type PatternRule struct {
	ID           string
	LanguageCode string // short code with optional variant
	Tokens       []*PatternToken
	Description  string
	Message      string
	ShortMessage string
	Tags         []rules.Tag
}

func NewPatternRule(id, languageCode string, tokens []*PatternToken, description, message, shortMessage string) *PatternRule {
	return &PatternRule{
		ID:           id,
		LanguageCode: languageCode,
		Tokens:       append([]*PatternToken(nil), tokens...),
		Description:  description,
		Message:      message,
		ShortMessage: shortMessage,
	}
}

func (r *PatternRule) GetID() string          { return r.ID }
func (r *PatternRule) GetDescription() string { return r.Description }
func (r *PatternRule) GetMessage() string     { return r.Message }
func (r *PatternRule) GetTags() []rules.Tag   { return r.Tags }
func (r *PatternRule) SetTags(tags []rules.Tag) {
	r.Tags = append([]rules.Tag(nil), tags...)
}
func (r *PatternRule) HasTag(tag rules.Tag) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// FalseFriendPatternRule ports org.languagetool.rules.patterns.FalseFriendPatternRule.
type FalseFriendPatternRule struct {
	*PatternRule
}

func NewFalseFriendPatternRule(id, languageCode string, tokens []*PatternToken, description, message, shortMessage string) *FalseFriendPatternRule {
	pr := NewPatternRule(id, languageCode, tokens, description, message, shortMessage)
	pr.SetTags([]rules.Tag{rules.TagPicky})
	return &FalseFriendPatternRule{PatternRule: pr}
}

// Match runs a simplified PatternRuleMatcher against the sentence.
func (r *PatternRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	return NewPatternRuleMatcherFromPattern(r).Match(sentence)
}
