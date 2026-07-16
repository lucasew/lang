package rules

// FakeRule ports org.languagetool.rules.FakeRule for unit tests.
type FakeRule struct {
	id string
}

func NewFakeRule(id string) *FakeRule {
	if id == "" {
		id = "FAKE-RULE"
	}
	return &FakeRule{id: id}
}

func (r *FakeRule) GetID() string { return r.id }

// PatternRule is a minimal stub of org.languagetool.rules.patterns.PatternRule
// sufficient for RuleWithMaxFilter / SameRuleGroupFilter unit tests.
type PatternRule struct {
	id    string
	subID *string
}

func NewPatternRule(id string) *PatternRule {
	return &PatternRule{id: id}
}

func (r *PatternRule) GetID() string { return r.id }

// GetSubID returns the sub-id pointer (nil = Java null).
func (r *PatternRule) GetSubID() *string { return r.subID }

func (r *PatternRule) SetSubID(sub string) {
	r.subID = &sub
}

// RuleWithID is implemented by rules that expose an identifier.
type RuleWithID interface {
	GetID() string
}

// AbstractPatternRule is the subset needed by RuleWithMaxFilter.haveSameRule.
type AbstractPatternRule interface {
	RuleWithID
	GetSubID() *string
}
