package diff

import "fmt"

// DiffStatus ports RuleMatchDiff.Status.
type DiffStatus string

const (
	DiffAdded    DiffStatus = "ADDED"
	DiffRemoved  DiffStatus = "REMOVED"
	DiffModified DiffStatus = "MODIFIED"
)

// RuleMatchDiff ports org.languagetool.dev.diff.RuleMatchDiff.
type RuleMatchDiff struct {
	Status   DiffStatus
	OldMatch *LightRuleMatch
	NewMatch *LightRuleMatch
}

func DiffAddedMatch(m *LightRuleMatch) RuleMatchDiff {
	return RuleMatchDiff{Status: DiffAdded, NewMatch: m}
}
func DiffRemovedMatch(m *LightRuleMatch) RuleMatchDiff {
	return RuleMatchDiff{Status: DiffRemoved, OldMatch: m}
}
func DiffModifiedMatch(oldM, newM *LightRuleMatch) RuleMatchDiff {
	return RuleMatchDiff{Status: DiffModified, OldMatch: oldM, NewMatch: newM}
}

func (d RuleMatchDiff) String() string {
	return fmt.Sprintf("%s: oldMatch=%v, newMatch=%v", d.Status, d.OldMatch, d.NewMatch)
}
