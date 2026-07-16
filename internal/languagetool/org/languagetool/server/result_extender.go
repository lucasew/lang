package server

// ExtensionMatch is a lightweight match used by ResultExtender (avoids rules import).
type ExtensionMatch struct {
	From        int
	To          int
	CategoryID  string
	CategoryName string
	IssueType   string
	Tags        []string
	EstimatedContext int
}

// HiddenMatch is the filtered hidden-rule representation.
type HiddenMatch struct {
	From             int
	To               int
	CategoryID       string
	CategoryName     string
	IssueType        string
	Tags             []string
	EstimatedContext int
	RuleID           string
	Message          string
}

const HiddenRuleID = "HIDDEN_RULE"

// GetAsHiddenMatches ports ResultExtender.getAsHiddenMatches:
// keep extension matches that do not touch any of the primary matches.
func GetAsHiddenMatches(matches, extensionMatches []ExtensionMatch) []HiddenMatch {
	var out []HiddenMatch
	for _, ext := range extensionMatches {
		touched := false
		for _, m := range matches {
			if ext.From <= m.To && ext.To >= m.From {
				touched = true
				break
			}
		}
		if touched {
			continue
		}
		issue := ext.IssueType
		if issue == "" {
			issue = "uncategorized"
		}
		out = append(out, HiddenMatch{
			From:             ext.From,
			To:               ext.To,
			CategoryID:       ext.CategoryID,
			CategoryName:     ext.CategoryName,
			IssueType:        issue,
			Tags:             ext.Tags,
			EstimatedContext: ext.EstimatedContext,
			RuleID:           HiddenRuleID,
			Message:          "(hidden message)",
		})
	}
	return out
}
