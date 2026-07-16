package wikipedia

// AppliedRuleMatch ports org.languagetool.dev.wikipedia.AppliedRuleMatch.
// RuleMatch identity is optional (from/to + message for green tests).
type AppliedRuleMatch struct {
	FromPos, ToPos int
	Message        string
	RuleMatchApps  []*RuleMatchApplication
}

func NewAppliedRuleMatch(from, to int, message string, apps []*RuleMatchApplication) *AppliedRuleMatch {
	return &AppliedRuleMatch{
		FromPos:       from,
		ToPos:         to,
		Message:       message,
		RuleMatchApps: apps,
	}
}

func (a *AppliedRuleMatch) GetRuleMatchApplications() []*RuleMatchApplication {
	if a == nil {
		return nil
	}
	return a.RuleMatchApps
}
