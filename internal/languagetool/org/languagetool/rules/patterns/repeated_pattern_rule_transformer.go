package patterns

import "strings"

// RepeatedPatternRuleTransformer ports grouping logic of
// org.languagetool.rules.patterns.RepeatedPatternRuleTransformer
// (full text-level match deferred; groups rules for wrapping).
type RepeatedPatternRuleTransformer struct {
	// DefaultMaxDistanceTokens ports defaultMaxDistanceTokens (token gap).
	DefaultMaxDistanceTokens int
	LanguageCode             string
}

func NewRepeatedPatternRuleTransformer(languageCode string) *RepeatedPatternRuleTransformer {
	return &RepeatedPatternRuleTransformer{
		DefaultMaxDistanceTokens: 60,
		LanguageCode:             languageCode,
	}
}

// Transform groups pattern rules that share the same id into remaining singles
// and transformed "repeated" groups (as TransformedRules metadata wrappers).
func (t *RepeatedPatternRuleTransformer) Transform(rules []*AbstractPatternRule) TransformedRules {
	if len(rules) == 0 {
		return NewTransformedRules(nil, nil)
	}
	byID := map[string][]*AbstractPatternRule{}
	var order []string
	for _, r := range rules {
		if r == nil {
			continue
		}
		if _, ok := byID[r.ID]; !ok {
			order = append(order, r.ID)
		}
		byID[r.ID] = append(byID[r.ID], r)
	}
	var remaining []*AbstractPatternRule
	var transformed []any
	for _, id := range order {
		group := byID[id]
		// only multi-member groups (or those marked with distance) become text-level
		wantTransform := len(group) > 1
		if !wantTransform {
			for _, r := range group {
				if r.DistanceTokens > 0 {
					wantTransform = true
					break
				}
			}
		}
		if wantTransform {
			transformed = append(transformed, &RepeatedPatternRuleGroup{
				ID:           id,
				LanguageCode: t.LanguageCode,
				Rules:        group,
				MaxDistance:  t.maxDistance(group),
			})
		} else {
			remaining = append(remaining, group...)
		}
	}
	return NewTransformedRules(remaining, transformed)
}

func (t *RepeatedPatternRuleTransformer) maxDistance(group []*AbstractPatternRule) int {
	max := t.DefaultMaxDistanceTokens
	for _, r := range group {
		if r.DistanceTokens > max {
			max = r.DistanceTokens
		}
	}
	return max
}

// RepeatedPatternRuleGroup is a lightweight stand-in for RepeatedPatternRule.
type RepeatedPatternRuleGroup struct {
	ID           string
	LanguageCode string
	Rules        []*AbstractPatternRule
	MaxDistance  int
}

func (g *RepeatedPatternRuleGroup) GetID() string {
	if g == nil || len(g.Rules) == 0 {
		return ""
	}
	return g.Rules[0].ID
}

func (g *RepeatedPatternRuleGroup) GetDescription() string {
	if g == nil || len(g.Rules) == 0 {
		return ""
	}
	return g.Rules[0].Description
}

func (g *RepeatedPatternRuleGroup) GetWrappedRules() []*AbstractPatternRule {
	return g.Rules
}

// ConsistencyPatternRuleTransformer ports id feature grouping helpers.
type ConsistencyPatternRuleTransformer struct {
	LanguageCode string
}

func NewConsistencyPatternRuleTransformer(languageCode string) *ConsistencyPatternRuleTransformer {
	return &ConsistencyPatternRuleTransformer{LanguageCode: languageCode}
}

// GetMainRuleId strips trailing _FEATURE style suffixes used by consistency rules.
// Java: id up to last underscore-separated feature marker; we take prefix before last '_'.
func GetMainRuleId(id string) string {
	if i := strings.LastIndex(id, "_"); i > 0 {
		return id[:i]
	}
	return id
}

// GetFeature returns the feature suffix after the main rule id.
func GetFeature(id string) string {
	if i := strings.LastIndex(id, "_"); i > 0 && i+1 < len(id) {
		return id[i+1:]
	}
	return ""
}

// Transform groups by main rule id.
func (t *ConsistencyPatternRuleTransformer) Transform(rules []*AbstractPatternRule) TransformedRules {
	byMain := map[string][]*AbstractPatternRule{}
	var order []string
	for _, r := range rules {
		if r == nil {
			continue
		}
		main := GetMainRuleId(r.ID)
		if _, ok := byMain[main]; !ok {
			order = append(order, main)
		}
		byMain[main] = append(byMain[main], r)
	}
	var remaining []*AbstractPatternRule
	var transformed []any
	for _, main := range order {
		group := byMain[main]
		if len(group) > 1 {
			transformed = append(transformed, &ConsistencyPatternRuleGroup{
				MainID:       main,
				LanguageCode: t.LanguageCode,
				Rules:        group,
			})
		} else {
			remaining = append(remaining, group...)
		}
	}
	return NewTransformedRules(remaining, transformed)
}

// ConsistencyPatternRuleGroup wraps related consistency pattern rules.
type ConsistencyPatternRuleGroup struct {
	MainID       string
	LanguageCode string
	Rules        []*AbstractPatternRule
}

func (g *ConsistencyPatternRuleGroup) GetID() string                           { return g.MainID }
func (g *ConsistencyPatternRuleGroup) GetWrappedRules() []*AbstractPatternRule { return g.Rules }
func (g *ConsistencyPatternRuleGroup) GetDescription() string {
	if len(g.Rules) == 0 {
		return ""
	}
	return g.Rules[0].Description
}
