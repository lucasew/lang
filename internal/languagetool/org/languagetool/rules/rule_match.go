package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// RuleMatch ports org.languagetool.rules.RuleMatch (fields needed by unit tests).
type RuleMatch struct {
	Rule                  any
	Sentence              *languagetool.AnalyzedSentence
	FromPos               int
	ToPos                 int
	Message               string
	ShortMessage          string
	SuggestedReplacements []string
}

func NewRuleMatch(rule any, sentence *languagetool.AnalyzedSentence, fromPos, toPos int, message string) *RuleMatch {
	return &RuleMatch{
		Rule:     rule,
		Sentence: sentence,
		FromPos:  fromPos,
		ToPos:    toPos,
		Message:  message,
	}
}

func (m *RuleMatch) GetFromPos() int { return m.FromPos }
func (m *RuleMatch) GetToPos() int   { return m.ToPos }
func (m *RuleMatch) SetSuggestedReplacement(s string) {
	m.SuggestedReplacements = []string{s}
}
func (m *RuleMatch) GetSuggestedReplacements() []string { return m.SuggestedReplacements }
