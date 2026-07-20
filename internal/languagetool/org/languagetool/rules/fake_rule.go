package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// Tag ports org.languagetool.Tag (subset used by filters).
type Tag string

const TagPicky Tag = "picky"

// FakeRule ports org.languagetool.rules.FakeRule for unit tests.
type FakeRule struct {
	id                string
	tags              []Tag
	includedAllAtOnce bool
	url               string
	toneTags          []languagetool.ToneTag
	goalSpecific      bool
	defaultTempOff    bool
}

func NewFakeRule(id string) *FakeRule {
	if id == "" {
		id = "FAKE-RULE"
	}
	return &FakeRule{id: id}
}

func NewFakeRuleWithTag(id string, tag Tag) *FakeRule {
	r := NewFakeRule(id)
	r.tags = []Tag{tag}
	return r
}

func (r *FakeRule) GetID() string { return r.id }

func (r *FakeRule) SetURL(u string) {
	if r != nil {
		r.url = u
	}
}

func (r *FakeRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.url
}

func (r *FakeRule) GetTags() []Tag {
	if r == nil || len(r.tags) == 0 {
		return nil
	}
	return append([]Tag(nil), r.tags...)
}

// SetTags ports Rule.setTags (used by FromLocalMatches + tests).
func (r *FakeRule) SetTags(tags []Tag) {
	if r == nil {
		return
	}
	if len(tags) == 0 {
		r.tags = nil
		return
	}
	r.tags = append([]Tag(nil), tags...)
}

func (r *FakeRule) HasTag(tag Tag) bool {
	for _, t := range r.tags {
		if t == tag {
			return true
		}
	}
	return false
}

func (r *FakeRule) IsIncludedInErrorsCorrectedAllAtOnce() bool {
	return r.includedAllAtOnce
}

// SetIncludedInErrorsCorrectedAllAtOnce ports PatternRule-style all-at-once flag for tests.
func (r *FakeRule) SetIncludedInErrorsCorrectedAllAtOnce(v bool) {
	if r != nil {
		r.includedAllAtOnce = v
	}
}

// GetToneTags ports Rule.getToneTags.
func (r *FakeRule) GetToneTags() []languagetool.ToneTag {
	if r == nil || len(r.toneTags) == 0 {
		return nil
	}
	return append([]languagetool.ToneTag(nil), r.toneTags...)
}

// SetToneTags ports Rule.setToneTags.
func (r *FakeRule) SetToneTags(tags ...languagetool.ToneTag) {
	if r != nil {
		r.toneTags = append([]languagetool.ToneTag(nil), tags...)
	}
}

// IsGoalSpecific ports Rule.isGoalSpecific.
func (r *FakeRule) IsGoalSpecific() bool {
	return r != nil && r.goalSpecific
}

// SetGoalSpecific ports Rule.setGoalSpecific.
func (r *FakeRule) SetGoalSpecific(v bool) {
	if r != nil {
		r.goalSpecific = v
	}
}

// IsDefaultTempOff ports Rule.isDefaultTempOff.
func (r *FakeRule) IsDefaultTempOff() bool {
	return r != nil && r.defaultTempOff
}

// SetDefaultTempOff ports Rule.setDefaultTempOff.
func (r *FakeRule) SetDefaultTempOff() {
	if r != nil {
		r.defaultTempOff = true
	}
}

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

// RuleWithTags optional tags for CleanOverlappingFilter.
type RuleWithTags interface {
	RuleWithID
	GetTags() []Tag
	IsIncludedInErrorsCorrectedAllAtOnce() bool
}
