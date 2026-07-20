package languagetool

// RuleLevelToneMeta is the rule surface needed by isRuleActiveForLevelAndToneTags
// without importing the rules package.
type RuleLevelToneMeta struct {
	IsPicky      bool // rule.hasTag(Tag.picky)
	ToneTags     []ToneTag
	GoalSpecific bool // rule.isGoalSpecific()
}

// IsRuleActiveForLevelAndToneTags ports JLanguageTool.isRuleActiveForLevelAndToneTags.
func IsRuleActiveForLevelAndToneTags(rule RuleLevelToneMeta, level Level, toneTags map[ToneTag]struct{}) bool {
	if level == LevelDefault && rule.IsPicky {
		return false
	}
	enabled := enabledToneTagsForFilter(toneTags)
	if len(rule.ToneTags) == 0 {
		return true
	}
	if containsTone(enabled, ToneAllWithoutGoalSpecific) {
		return !rule.GoalSpecific
	}
	for _, t := range enabled {
		for _, rt := range rule.ToneTags {
			if rt == t {
				return true
			}
		}
	}
	return false
}

// enabledToneTagsForFilter ports the enabledToneTags list construction in
// isRuleActiveForLevelAndToneTags.
func enabledToneTagsForFilter(toneTags map[ToneTag]struct{}) []ToneTag {
	if _, ok := toneTags[ToneAllToneRules]; ok {
		return RealToneTags()
	}
	if _, ok := toneTags[ToneNoToneRule]; ok {
		return nil // even clarity disabled
	}
	if len(toneTags) == 0 {
		return []ToneTag{ToneAllWithoutGoalSpecific}
	}
	if _, ok := toneTags[ToneAllWithoutGoalSpecific]; ok {
		return []ToneTag{ToneAllWithoutGoalSpecific}
	}
	out := make([]ToneTag, 0, len(toneTags))
	for t := range toneTags {
		out = append(out, t)
	}
	return out
}

func containsTone(list []ToneTag, want ToneTag) bool {
	for _, t := range list {
		if t == want {
			return true
		}
	}
	return false
}

// LocalMatchLevelToneMeta builds RuleLevelToneMeta from a LocalMatch
// (IsPicky, ToneTags, GoalSpecific from ToLocalMatches / injects).
func LocalMatchLevelToneMeta(m LocalMatch) RuleLevelToneMeta {
	return RuleLevelToneMeta{
		IsPicky:      m.IsPicky,
		ToneTags:     append([]ToneTag(nil), m.ToneTags...),
		GoalSpecific: m.GoalSpecific,
	}
}

// FilterMatchesForLevelAndToneTags ports filterMatches tone/level stream filter.
func FilterMatchesForLevelAndToneTags(ms []LocalMatch, level Level, toneTags map[ToneTag]struct{}) []LocalMatch {
	if len(ms) == 0 {
		return ms
	}
	out := make([]LocalMatch, 0, len(ms))
	for _, m := range ms {
		if IsRuleActiveForLevelAndToneTags(LocalMatchLevelToneMeta(m), level, toneTags) {
			out = append(out, m)
		}
	}
	return out
}
