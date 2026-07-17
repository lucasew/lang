package languagetool

import "strings"

// FilterMatchesByCategories drops matches whose SoftRuleMeta / LocalMatch category
// is disabled, or not in the enabled category set when any are listed.
// Soft: --enablecategories alone restricts to those categories (does not require
// --enabledonly, which is reserved for rule-id enable-only mode).
func FilterMatchesByCategories(ms []LocalMatch, disabled, enabled []string, enabledOnly bool) []LocalMatch {
	if len(ms) == 0 {
		return ms
	}
	dis := foldSet(disabled)
	en := foldSet(enabled)
	// enabledOnly is accepted for API parity; non-empty enabled always restricts.
	restrictToEnabled := len(en) > 0
	if len(dis) == 0 && !restrictToEnabled {
		return ms
	}
	_ = enabledOnly
	out := make([]LocalMatch, 0, len(ms))
	for _, m := range ms {
		catID := m.CategoryID
		if catID == "" {
			catID, _, _, _ = SoftRuleMeta(m.RuleID)
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
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		m[strings.ToUpper(s)] = struct{}{}
	}
	return m
}
