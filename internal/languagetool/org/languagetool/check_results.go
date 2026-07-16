package languagetool

import "sort"

// CheckResults ports org.languagetool.CheckResults.
// RuleMatches uses any to avoid importing rules (call sites use []*rules.RuleMatch).
type CheckResults struct {
	RuleMatches            []any
	IgnoredRanges          []Range
	ExtendedSentenceRanges []ExtendedSentenceRange
	sentenceRanges         []SentenceRange
}

func NewCheckResults(ruleMatches []any, ignoredRanges []Range) *CheckResults {
	return NewCheckResultsFull(ruleMatches, ignoredRanges, nil)
}

func NewCheckResultsFull(ruleMatches []any, ignoredRanges []Range, extended []ExtendedSentenceRange) *CheckResults {
	ext := append([]ExtendedSentenceRange(nil), extended...)
	sort.Slice(ext, func(i, j int) bool { return ext[i].Less(ext[j]) })
	return &CheckResults{
		RuleMatches:            ruleMatches,
		IgnoredRanges:          append([]Range(nil), ignoredRanges...),
		ExtendedSentenceRanges: ext,
	}
}

func (c *CheckResults) GetRuleMatches() []any { return c.RuleMatches }
func (c *CheckResults) SetRuleMatches(m []any) {
	c.RuleMatches = m
}
func (c *CheckResults) GetIgnoredRanges() []Range { return c.IgnoredRanges }
func (c *CheckResults) GetExtendedSentenceRanges() []ExtendedSentenceRange {
	return c.ExtendedSentenceRanges
}
func (c *CheckResults) GetSentenceRanges() []SentenceRange {
	return append([]SentenceRange(nil), c.sentenceRanges...)
}
func (c *CheckResults) AddSentenceRanges(ranges []SentenceRange) {
	c.sentenceRanges = append(c.sentenceRanges, ranges...)
}
