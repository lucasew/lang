package commandline

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// FilterMatchesByRules applies enable/disable lists to matches.
//
// Java Tools.selectRules: non-empty enabledRules calls enableRule; only when
// useEnabledOnly does it disable all rules except those listed. Enabling rules
// alone must not hide matches from other active rules.
func FilterMatchesByRules(matches []*rules.RuleMatch, disabled, enabled []string, enabledOnly bool) []*rules.RuleMatch {
	dis := toSet(disabled)
	en := toSet(enabled)
	// Java: restrict to enabled rule IDs only with useEnabledOnly.
	restrictToEnabled := enabledOnly && len(en) > 0
	var out []*rules.RuleMatch
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		if _, ok := dis[id]; ok {
			continue
		}
		if restrictToEnabled {
			if _, ok := en[id]; !ok {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func toSet(ss []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, s := range ss {
		if s != "" {
			m[s] = struct{}{}
		}
	}
	return m
}
