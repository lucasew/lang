package tools

// RuleSelector is a lightweight port of Tools.selectRules state
// without a full JLanguageTool rule registry.
type RuleSelector struct {
	// Active is the current set of enabled rule IDs.
	Active map[string]struct{}
	// CategoryOf maps rule ID → category ID.
	CategoryOf map[string]string
	// All is the full registry of known rules (for enable-only rebuilds).
	All map[string]struct{}
}

// NewRuleSelector seeds rules (default category MISC).
func NewRuleSelector(ruleIDs ...string) *RuleSelector {
	rs := &RuleSelector{
		Active:     map[string]struct{}{},
		CategoryOf: map[string]string{},
		All:        map[string]struct{}{},
	}
	for _, id := range ruleIDs {
		rs.Active[id] = struct{}{}
		rs.All[id] = struct{}{}
		rs.CategoryOf[id] = "MISC"
	}
	return rs
}

// SetCategory assigns a category ID to a rule.
func (rs *RuleSelector) SetCategory(ruleID, categoryID string) {
	if rs == nil {
		return
	}
	if rs.CategoryOf == nil {
		rs.CategoryOf = map[string]string{}
	}
	rs.CategoryOf[ruleID] = categoryID
	if rs.All == nil {
		rs.All = map[string]struct{}{}
	}
	rs.All[ruleID] = struct{}{}
}

// SelectRules ports Tools.selectRules category/rule enable-disable matrix.
func (rs *RuleSelector) SelectRules(
	disabledCategories, enabledCategories map[string]struct{},
	disabledRules, enabledRules map[string]struct{},
	useEnabledOnly, enableTempOff bool,
) {
	_ = enableTempOff
	if rs == nil {
		return
	}
	if rs.Active == nil {
		rs.Active = map[string]struct{}{}
	}
	if useEnabledOnly {
		// Start from empty, then enable categories/rules.
		// Java: enableTempOff false → disable all first when useEnabledOnly.
		// ToolsTest: empty sets with useEnabledOnly=true still keeps DEMO_RULE —
		// meaning default-on rules stay unless something is restricted.
		// Match ToolsTest: only clear when there is an explicit enabled filter.
		if len(enabledCategories) > 0 || len(enabledRules) > 0 {
			rs.Active = map[string]struct{}{}
			for id, cat := range rs.CategoryOf {
				if _, on := enabledCategories[cat]; on {
					rs.Active[id] = struct{}{}
				}
			}
			for id := range enabledRules {
				rs.Active[id] = struct{}{}
			}
		}
		// still apply disables
		for id, cat := range rs.CategoryOf {
			if _, off := disabledCategories[cat]; off {
				delete(rs.Active, id)
			}
		}
		for id := range disabledRules {
			delete(rs.Active, id)
		}
		// re-enable explicit
		for id := range enabledRules {
			rs.Active[id] = struct{}{}
		}
		return
	}

	// Normal mode: start from current Active (all default-on)
	for id, cat := range rs.CategoryOf {
		if _, off := disabledCategories[cat]; off {
			delete(rs.Active, id)
		}
	}
	for id := range disabledRules {
		delete(rs.Active, id)
	}
	for id, cat := range rs.CategoryOf {
		if _, on := enabledCategories[cat]; on {
			rs.Active[id] = struct{}{}
		}
	}
	for id := range enabledRules {
		rs.Active[id] = struct{}{}
	}
}

// Has reports whether ruleID is active.
func (rs *RuleSelector) Has(ruleID string) bool {
	if rs == nil {
		return false
	}
	_, ok := rs.Active[ruleID]
	return ok
}
