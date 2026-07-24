package rules

// LanguageDependentRuleMatchFilter ports
// org.languagetool.rules.LanguageDependentRuleMatchFilter.
type LanguageDependentRuleMatchFilter struct {
	// Filter is the language-specific match filter (text is plain/annotated text).
	Filter func(matches []*RuleMatch, text string, enabledRuleIDs map[string]struct{}) []*RuleMatch
	// EnabledRuleIDs is the active rule set; nil means no filtering by id.
	EnabledRuleIDs map[string]struct{}
}

func NewLanguageDependentRuleMatchFilter(filter func([]*RuleMatch, string, map[string]struct{}) []*RuleMatch, enabled []string) *LanguageDependentRuleMatchFilter {
	m := map[string]struct{}{}
	for _, id := range enabled {
		m[id] = struct{}{}
	}
	return &LanguageDependentRuleMatchFilter{Filter: filter, EnabledRuleIDs: m}
}

// Apply runs the language filter, or returns matches unchanged when Filter is nil.
func (f *LanguageDependentRuleMatchFilter) Apply(matches []*RuleMatch, text string) []*RuleMatch {
	if f.Filter == nil {
		return matches
	}
	return f.Filter(matches, text, f.EnabledRuleIDs)
}
