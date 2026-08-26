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
	// DefaultTempOff ports Rule.isDefaultTempOff — enabled only when enableTempOff.
	DefaultTempOff map[string]struct{}
}

// NewRuleSelector seeds rules (default category MISC).
func NewRuleSelector(ruleIDs ...string) *RuleSelector {
	rs := &RuleSelector{
		Active:         map[string]struct{}{},
		CategoryOf:     map[string]string{},
		All:            map[string]struct{}{},
		DefaultTempOff: map[string]struct{}{},
	}
	for _, id := range ruleIDs {
		rs.Active[id] = struct{}{}
		rs.All[id] = struct{}{}
		rs.CategoryOf[id] = "MISC"
	}
	return rs
}

// SetDefaultTempOff marks a rule as default temp_off (Java Rule.isDefaultTempOff).
func (rs *RuleSelector) SetDefaultTempOff(ruleID string) {
	if rs == nil {
		return
	}
	if rs.DefaultTempOff == nil {
		rs.DefaultTempOff = map[string]struct{}{}
	}
	if rs.All == nil {
		rs.All = map[string]struct{}{}
	}
	rs.DefaultTempOff[ruleID] = struct{}{}
	rs.All[ruleID] = struct{}{}
	if rs.CategoryOf == nil {
		rs.CategoryOf = map[string]string{}
	}
	if _, ok := rs.CategoryOf[ruleID]; !ok {
		rs.CategoryOf[ruleID] = "MISC"
	}
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
// Order matches Java: enableTempOff → disable categories → enable categories
// (+ useEnabledOnly category filter) → disable rules → enable rules
// (+ useEnabledOnly rule filter).
func (rs *RuleSelector) SelectRules(
	disabledCategories, enabledCategories map[string]struct{},
	disabledRules, enabledRules map[string]struct{},
	useEnabledOnly, enableTempOff bool,
) {
	if rs == nil {
		return
	}
	if rs.Active == nil {
		rs.Active = map[string]struct{}{}
	}
	// Java: if (enableTempOff) enable default temp_off rules first.
	if enableTempOff {
		for id := range rs.DefaultTempOff {
			rs.Active[id] = struct{}{}
		}
	}
	// disable categories
	for id, cat := range rs.CategoryOf {
		if _, off := disabledCategories[cat]; off {
			delete(rs.Active, id)
		}
	}
	// enable categories (+ useEnabledOnly: disable rules outside enabled categories)
	if len(enabledCategories) > 0 {
		for id, cat := range rs.CategoryOf {
			if _, on := enabledCategories[cat]; on {
				rs.Active[id] = struct{}{}
			}
		}
		if useEnabledOnly {
			for id, cat := range rs.CategoryOf {
				if _, on := enabledCategories[cat]; !on {
					delete(rs.Active, id)
				}
			}
		}
	}
	// disable rules explicitly
	for id := range disabledRules {
		delete(rs.Active, id)
	}
	// enable rules (+ useEnabledOnly: disable rules not in enabled set)
	if len(enabledRules) > 0 {
		for id := range enabledRules {
			rs.Active[id] = struct{}{}
		}
		if useEnabledOnly {
			// Java: disable unless enabledRules contains fullId or id.
			// RuleSelector tracks id only (no sub-ids).
			for id := range rs.All {
				if _, on := enabledRules[id]; !on {
					delete(rs.Active, id)
				}
			}
			// re-assert enabled (All may not list an ad-hoc id)
			for id := range enabledRules {
				rs.Active[id] = struct{}{}
			}
		}
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
