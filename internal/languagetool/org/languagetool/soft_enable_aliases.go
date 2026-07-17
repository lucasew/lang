package languagetool

import "strings"

// ExpandSoftEnableRuleIDs expands soft bulk-enable tokens among enabled rule IDs.
// SOFT_OPTIONAL / SOFT_OPT_ALL → every registered id containing SOFT_OPT_
// (optional soft packs that start with default="off").
func ExpandSoftEnableRuleIDs(registered, enabled []string) []string {
	if len(enabled) == 0 {
		return enabled
	}
	var out []string
	seen := map[string]struct{}{}
	add := func(id string) {
		if id == "" {
			return
		}
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	for _, id := range enabled {
		up := strings.ToUpper(strings.TrimSpace(id))
		if up == "SOFT_OPTIONAL" || up == "SOFT_OPT_ALL" {
			for _, rid := range registered {
				if strings.Contains(rid, "SOFT_OPT_") {
					add(rid)
				}
			}
			continue
		}
		add(id)
	}
	return out
}
