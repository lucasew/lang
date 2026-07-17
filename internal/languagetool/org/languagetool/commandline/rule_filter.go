package commandline

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// FilterMatchesByRules applies enable/disable lists to matches.
// disabled: drop if ID in set
// enabledOnly: if non-empty, keep only IDs in set
func FilterMatchesByRules(matches []*rules.RuleMatch, disabled, enabled []string, enabledOnly bool) []*rules.RuleMatch {
	dis := toSet(disabled)
	en := toSet(enabled)
	var out []*rules.RuleMatch
	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		if _, ok := dis[id]; ok {
			continue
		}
		if enabledOnly || len(en) > 0 {
			// if enabled list non-empty without enabledOnly, Java still only enables listed when -e used with all others on...
			// Green: when enabled set non-empty, keep only those
			if len(en) > 0 {
				if _, ok := en[id]; !ok {
					continue
				}
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
