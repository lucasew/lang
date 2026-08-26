package diff

import (
	"fmt"
	"regexp"
	"strings"
)

// MatchStatus ports LightRuleMatch.Status.
type MatchStatus string

const (
	StatusOn     MatchStatus = "on"
	StatusTempOff MatchStatus = "temp_off"
)

// LightRuleMatch ports org.languagetool.dev.diff.LightRuleMatch.
type LightRuleMatch struct {
	Line        int
	Column      int
	FullRuleID  string
	Message     string
	Category    string
	Context     string
	CoveredText string
	Suggestions []string
	RuleSource  string
	Title       string
	Status      MatchStatus
	Tags        []string
	Premium     bool
}

var subIDRE = regexp.MustCompile(`\[(\d+)\]`)

func MasterRuleID(full string) string {
	return strings.TrimSpace(subIDRE.ReplaceAllString(full, ""))
}

func SubRuleID(full string) string {
	m := subIDRE.FindStringSubmatch(full)
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

func (m *LightRuleMatch) GetRuleID() string {
	if m == nil {
		return ""
	}
	return MasterRuleID(m.FullRuleID)
}

func (m *LightRuleMatch) GetSubID() string {
	if m == nil {
		return ""
	}
	return SubRuleID(m.FullRuleID)
}

func (m *LightRuleMatch) GetFullRuleID() string {
	if m == nil {
		return ""
	}
	return m.FullRuleID
}

func (m *LightRuleMatch) String() string {
	if m == nil {
		return "null"
	}
	sub := m.GetSubID()
	if sub == "" {
		sub = "null"
	}
	return fmt.Sprintf("%d/%d %s[%s], msg=%s, covered=%s, suggestions=%s, title=%s, ctx=%s",
		m.Line, m.Column, m.GetRuleID(), sub, m.Message, m.CoveredText,
		fmt.Sprint(m.Suggestions), m.Title, m.Context)
}
