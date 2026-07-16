package server

import "fmt"

// Span is a minimal stand-in for a RuleMatch span used by overlap checks.
type Span struct {
	From int
	To   int
}

// RemoteRuleMatch ports org.languagetool.server.RemoteRuleMatch.
type RemoteRuleMatch struct {
	RuleID                      string
	Message                     string
	Context                     string
	ContextOffset               int
	Offset                      int
	ErrorLength                 int
	EstimatedContextForSureMatch int

	SubID               string
	ShortMessage        string
	Replacements        []string
	URL                 string
	Category            string
	CategoryID          string
	LocQualityIssueType string
	Tags                []string
}

func NewRemoteRuleMatch(ruleID, msg, context string, contextOffset, offset, errorLength int) *RemoteRuleMatch {
	return NewRemoteRuleMatchFull(ruleID, msg, context, contextOffset, offset, errorLength, 0)
}

func NewRemoteRuleMatchFull(ruleID, msg, context string, contextOffset, offset, errorLength, estimated int) *RemoteRuleMatch {
	if ruleID == "" || msg == "" || context == "" {
		panic("ruleId, msg, and context are required")
	}
	return &RemoteRuleMatch{
		RuleID:                       ruleID,
		Message:                      msg,
		Context:                      context,
		ContextOffset:                contextOffset,
		Offset:                       offset,
		ErrorLength:                  errorLength,
		EstimatedContextForSureMatch: estimated,
	}
}

// IsTouchedByOneOf reports overlap with any span (from/to exclusive end like Java getToPos).
func (m *RemoteRuleMatch) IsTouchedByOneOf(spans []Span) bool {
	if m == nil {
		return false
	}
	end := m.Offset + m.ErrorLength
	for _, s := range spans {
		if m.Offset <= s.To && end >= s.From {
			return true
		}
	}
	return false
}

func (m *RemoteRuleMatch) String() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("%s@%d-%d", m.RuleID, m.Offset, m.Offset+m.ErrorLength)
}

// ToMatchInfo converts to the public API MatchInfo shape.
func (m *RemoteRuleMatch) ToMatchInfo() MatchInfo {
	if m == nil {
		return MatchInfo{}
	}
	info := MatchInfo{
		Message:      m.Message,
		ShortMessage: m.ShortMessage,
		Offset:       m.Offset,
		Length:       m.ErrorLength,
		Context: ContextInfo{
			Text:   m.Context,
			Offset: m.ContextOffset,
			Length: m.ErrorLength,
		},
		Rule: RuleInfo{
			ID:          m.RuleID,
			Description: m.Message,
			IssueType:   m.LocQualityIssueType,
		},
	}
	info.Rule.Category.ID = m.CategoryID
	info.Rule.Category.Name = m.Category
	for _, r := range m.Replacements {
		info.Replacements = append(info.Replacements, ReplacementInfo{Value: r})
	}
	return info
}
