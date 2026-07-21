package languagetool

import "strings"

// FilterMatchesByCategories drops matches whose RuleMeta / LocalMatch category
// is disabled, or (when enabledOnly) not in the enabled category set.
//
// Java Tools.selectRules: non-empty enabledCategories calls enableRuleCategory;
// only when useEnabledOnly does it disable rules outside those categories.
// Enabling categories alone must not hide matches from other categories.
func FilterMatchesByCategories(ms []LocalMatch, disabled, enabled []string, enabledOnly bool) []LocalMatch {
	if len(ms) == 0 {
		return ms
	}
	dis := foldSet(disabled)
	en := foldSet(enabled)
	// Java: restrict to enabled categories only with useEnabledOnly.
	restrictToEnabled := enabledOnly && len(en) > 0
	if len(dis) == 0 && !restrictToEnabled {
		return ms
	}
	out := make([]LocalMatch, 0, len(ms))
	for _, m := range ms {
		catID := m.CategoryID
		if catID == "" {
			catID, _, _, _ = RuleMeta(m.RuleID)
		}
		key := strings.ToUpper(catID)
		if _, drop := dis[key]; drop {
			continue
		}
		if restrictToEnabled {
			if _, ok := en[key]; !ok {
				continue
			}
		}
		out = append(out, m)
	}
	return out
}

func foldSet(items []string) map[string]struct{} {
	if len(items) == 0 {
		return nil
	}
	m := make(map[string]struct{}, len(items))
	for _, s := range items {
		// Java CommandLineParser / API: categories.split(",") — no per-item trim.
		// Keep surface as-is (case-fold only); skip pure empty CSV slots.
		if s == "" {
			continue
		}
		m[strings.ToUpper(s)] = struct{}{}
	}
	return m
}
