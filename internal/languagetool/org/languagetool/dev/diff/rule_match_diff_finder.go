package diff

// RuleMatchDiffFinder ports getDiffs core (HTML/report deferred).
type RuleMatchDiffFinder struct{}

func NewRuleMatchDiffFinder() *RuleMatchDiffFinder { return &RuleMatchDiffFinder{} }

// matchKey identifies "same" matches across runs.
type matchKey struct {
	line, column int
	ruleID       string
	title        string
	covered      string
}

// GetDiffs compares old vs new match lists.
func (f *RuleMatchDiffFinder) GetDiffs(oldL, newL []*LightRuleMatch) []RuleMatchDiff {
	oldMap := map[matchKey]*LightRuleMatch{}
	for _, m := range oldL {
		if m == nil {
			continue
		}
		oldMap[keyOf(m)] = m
	}
	newMap := map[matchKey]*LightRuleMatch{}
	for _, m := range newL {
		if m == nil {
			continue
		}
		newMap[keyOf(m)] = m
	}

	var result []RuleMatchDiff
	for _, m := range newL {
		if m == nil {
			continue
		}
		k := keyOf(m)
		if old, ok := oldMap[k]; ok {
			if !sameSuggestions(old.Suggestions, m.Suggestions) ||
				old.Message != m.Message ||
				old.Status != m.Status ||
				old.CoveredText != m.CoveredText {
				result = append(result, DiffModifiedMatch(old, m))
			}
		} else {
			result = append(result, DiffAddedMatch(m))
		}
	}
	for _, m := range oldL {
		if m == nil {
			continue
		}
		if _, ok := newMap[keyOf(m)]; !ok {
			result = append(result, DiffRemovedMatch(m))
		}
	}
	return result
}

func keyOf(m *LightRuleMatch) matchKey {
	return matchKey{
		line: m.Line, column: m.Column, ruleID: m.GetRuleID(),
		title: m.Title, covered: m.CoveredText,
	}
}

func sameSuggestions(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// DiffsString formats like Java List.toString for tests.
func DiffsString(diffs []RuleMatchDiff) string {
	if len(diffs) == 0 {
		return "[]"
	}
	s := "["
	for i, d := range diffs {
		if i > 0 {
			s += ", "
		}
		s += d.String()
	}
	s += "]"
	return s
}
