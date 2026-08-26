package remote

import "strconv"

// RemoteRuleMatch ports org.languagetool.remote.RemoteRuleMatch.
type RemoteRuleMatch struct {
	RuleID              string
	RuleDescription     string
	Message             string
	Context             string
	ContextOffset       int
	Offset              int
	ErrorLength         int
	SubID               string
	ShortMessage        string
	Replacements        []string
	URL                 string
	Category            string
	CategoryID          string
	LocQualityIssueType string
	// ContextForSureMatch soft-ports API contextForSureMatch.
	ContextForSureMatch int
	// TypeName soft-ports matches[].type.typeName.
	TypeName string
}

func NewRemoteRuleMatch(ruleID, ruleDescription, msg, context string, contextOffset, offset, errorLength int) *RemoteRuleMatch {
	if ruleID == "" || msg == "" || context == "" {
		panic("ruleID, msg, and context are required")
	}
	return &RemoteRuleMatch{
		RuleID:          ruleID,
		RuleDescription: ruleDescription,
		Message:         msg,
		Context:         context,
		ContextOffset:   contextOffset,
		Offset:          offset,
		ErrorLength:     errorLength,
	}
}

func (m *RemoteRuleMatch) GetRuleID() string          { return m.RuleID }
func (m *RemoteRuleMatch) GetRuleDescription() string { return m.RuleDescription }
func (m *RemoteRuleMatch) GetMessage() string         { return m.Message }
func (m *RemoteRuleMatch) GetContext() string         { return m.Context }
func (m *RemoteRuleMatch) GetContextOffset() int      { return m.ContextOffset }
func (m *RemoteRuleMatch) GetOffset() int             { return m.Offset }

// GetErrorOffset ports RemoteRuleMatch.getErrorOffset (Java name for offset).
func (m *RemoteRuleMatch) GetErrorOffset() int { return m.Offset }
func (m *RemoteRuleMatch) GetErrorLength() int { return m.ErrorLength }
func (m *RemoteRuleMatch) GetRuleSubID() (string, bool) {
	if m.SubID == "" {
		return "", false
	}
	return m.SubID, true
}
func (m *RemoteRuleMatch) GetShortMessage() (string, bool) {
	if m.ShortMessage == "" {
		return "", false
	}
	return m.ShortMessage, true
}

// GetReplacements ports Optional<List<String>> getReplacements.
// ok=false when unset (Java Optional.empty).
func (m *RemoteRuleMatch) GetReplacements() ([]string, bool) {
	if m == nil || m.Replacements == nil {
		return nil, false
	}
	return append([]string(nil), m.Replacements...), true
}
func (m *RemoteRuleMatch) SetReplacements(r []string) {
	m.Replacements = append([]string(nil), r...)
}
func (m *RemoteRuleMatch) SetSubID(id string)       { m.SubID = id }
func (m *RemoteRuleMatch) SetShortMessage(s string) { m.ShortMessage = s }
func (m *RemoteRuleMatch) SetURL(u string)          { m.URL = u }
func (m *RemoteRuleMatch) SetCategory(c, id string) { m.Category, m.CategoryID = c, id }
func (m *RemoteRuleMatch) SetLocQualityIssueType(t string) { m.LocQualityIssueType = t }
func (m *RemoteRuleMatch) SetContextForSureMatch(n int)    { m.ContextForSureMatch = n }
func (m *RemoteRuleMatch) SetTypeName(t string)            { m.TypeName = t }

// GetURL ports Optional<String> getUrl.
func (m *RemoteRuleMatch) GetURL() (string, bool) {
	if m == nil || m.URL == "" {
		return "", false
	}
	return m.URL, true
}

// GetCategory ports Optional<String> getCategory.
func (m *RemoteRuleMatch) GetCategory() (string, bool) {
	if m == nil || m.Category == "" {
		return "", false
	}
	return m.Category, true
}

// GetCategoryID ports Optional<String> getCategoryId.
func (m *RemoteRuleMatch) GetCategoryID() (string, bool) {
	if m == nil || m.CategoryID == "" {
		return "", false
	}
	return m.CategoryID, true
}

// GetLocQualityIssueType ports Optional<String> getLocQualityIssueType.
func (m *RemoteRuleMatch) GetLocQualityIssueType() (string, bool) {
	if m == nil || m.LocQualityIssueType == "" {
		return "", false
	}
	return m.LocQualityIssueType, true
}
func (m *RemoteRuleMatch) GetContextForSureMatch() int {
	if m == nil {
		return 0
	}
	return m.ContextForSureMatch
}
func (m *RemoteRuleMatch) GetTypeName() string {
	if m == nil {
		return ""
	}
	return m.TypeName
}

// String ports RemoteRuleMatch.toString → "ruleId@offset-(offset+len)".
func (m *RemoteRuleMatch) String() string {
	if m == nil {
		return ""
	}
	return m.RuleID + "@" + strconv.Itoa(m.Offset) + "-" + strconv.Itoa(m.Offset+m.ErrorLength)
}
