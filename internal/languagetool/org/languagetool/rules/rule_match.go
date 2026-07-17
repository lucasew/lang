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
	// URL optional match-level link (overrides rule URL when set).
	URL string
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

func (m *RuleMatch) GetRule() any      { return m.Rule }
func (m *RuleMatch) GetMessage() string { return m.Message }
func (m *RuleMatch) GetShortMessage() string {
	if m == nil {
		return ""
	}
	return m.ShortMessage
}

func (m *RuleMatch) SetOffsetPosition(from, to int) {
	m.FromPos = from
	m.ToPos = to
}

func (m *RuleMatch) SetSuggestedReplacements(reps []string) {
	m.SuggestedReplacements = append([]string(nil), reps...)
}

func (m *RuleMatch) SetURL(u string) {
	if m != nil {
		m.URL = u
	}
}

func (m *RuleMatch) GetURL() string {
	if m == nil {
		return ""
	}
	return m.URL
}
