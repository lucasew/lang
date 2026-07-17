package languagetool

import "strings"

// FilterMatchesByCategories drops matches whose SoftRuleMeta / LocalMatch category
// is disabled, or (when enabledOnly) not in the enabled category set.
func FilterMatchesByCategories(ms []LocalMatch, disabled, enabled []string, enabledOnly bool) []LocalMatch {
	if len(ms) == 0 {
		return ms
	}
	dis := foldSet(disabled)
	en := foldSet(enabled)
	if len(dis) == 0 && !(enabledOnly && len(en) > 0) {
		return ms
	}
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
		if enabledOnly && len(en) > 0 {
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
