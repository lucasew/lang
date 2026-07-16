package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// Rule is the minimal interface for sentence-level language rules
// (subset of org.languagetool.rules.Rule).
type Rule interface {
	GetID() string
	GetDescription() string
	Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch
}

// RuleWithError is used when Match can fail (I/O spellers).
type RuleWithError interface {
	GetID() string
	Match(sentence *languagetool.AnalyzedSentence) ([]*RuleMatch, error)
}

// BaseRule holds common metadata for concrete rules.
type BaseRule struct {
	ID          string
	Description string
	Category    *Category
	DefaultOff  bool
}

func (r *BaseRule) GetID() string {
	if r == nil {
		return ""
	}
	return r.ID
}

func (r *BaseRule) GetDescription() string {
	if r == nil {
		return ""
	}
	return r.Description
}

func (r *BaseRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

func (r *BaseRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *BaseRule) SetCategory(c *Category) {
	if r != nil {
		r.Category = c
	}
}
