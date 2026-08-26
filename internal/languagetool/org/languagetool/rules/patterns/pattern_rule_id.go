package patterns

// PatternRuleId ports org.languagetool.rules.patterns.PatternRuleId.
type PatternRuleId struct {
	ID    string
	SubID *string
}

func NewPatternRuleId(id string) PatternRuleId {
	if id == "" {
		panic("id must be set")
	}
	return PatternRuleId{ID: id}
}

func NewPatternRuleIdWithSub(id, subID string) PatternRuleId {
	if id == "" {
		panic("id must be set")
	}
	if subID == "" {
		panic("subId must be set, if specified")
	}
	s := subID
	return PatternRuleId{ID: id, SubID: &s}
}

func (p PatternRuleId) GetID() string { return p.ID }

// GetSubID returns the sub-id or empty if unset.
func (p PatternRuleId) GetSubID() *string { return p.SubID }

func (p PatternRuleId) String() string {
	if p.SubID != nil {
		return p.ID + "[" + *p.SubID + "]"
	}
	return p.ID
}
