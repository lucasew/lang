package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RepeatedPatternRuleTransformer ports
// org.languagetool.rules.patterns.RepeatedPatternRuleTransformer.
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

// Transform ports apply(): only rules with getMinPrevMatches() > 0 become
// RepeatedPatternRule groups (keyed by rule id); others remain sentence-level.
func (t *RepeatedPatternRuleTransformer) Transform(rules []*AbstractPatternRule) TransformedRules {
	if len(rules) == 0 {
		return NewTransformedRules(nil, nil)
	}
	toTransform := map[string][]*AbstractPatternRule{}
	var order []string
	var remaining []*AbstractPatternRule
	for _, r := range rules {
		if r == nil {
			continue
		}
		if r.GetMinPrevMatches() > 0 {
			if _, ok := toTransform[r.ID]; !ok {
				order = append(order, r.ID)
			}
			toTransform[r.ID] = append(toTransform[r.ID], r)
			continue
		}
		remaining = append(remaining, r)
	}
	var transformed []any
	for _, id := range order {
		group := toTransform[id]
		transformed = append(transformed, &RepeatedPatternRule{
			LanguageCode:             t.LanguageCode,
			AbstractRules:            group,
			DefaultMaxDistanceTokens: t.DefaultMaxDistanceTokens,
		})
	}
	return NewTransformedRules(remaining, transformed)
}

// RepeatedPatternRule ports RepeatedPatternRuleTransformer.RepeatedPatternRule
// (TextLevelRule wrapper around one or more AbstractPatternRule / PatternRule twins).
type RepeatedPatternRule struct {
	LanguageCode             string
	AbstractRules            []*AbstractPatternRule
	// PatternRules are concrete matchers (preferred when set by RegisterGrammar).
	PatternRules []*PatternRule
	// DefaultMaxDistanceTokens ports transformer defaultMaxDistanceTokens.
	DefaultMaxDistanceTokens int
}

func (r *RepeatedPatternRule) GetID() string {
	if r == nil {
		return ""
	}
	if len(r.PatternRules) > 0 && r.PatternRules[0] != nil {
		return r.PatternRules[0].GetID()
	}
	if len(r.AbstractRules) > 0 && r.AbstractRules[0] != nil {
		return r.AbstractRules[0].ID
	}
	return ""
}

func (r *RepeatedPatternRule) GetDescription() string {
	if r == nil {
		return ""
	}
	if len(r.PatternRules) > 0 && r.PatternRules[0] != nil {
		return r.PatternRules[0].GetDescription()
	}
	if len(r.AbstractRules) > 0 && r.AbstractRules[0] != nil {
		return r.AbstractRules[0].Description
	}
	return ""
}

func (r *RepeatedPatternRule) GetWrappedRules() []*AbstractPatternRule {
	if r == nil {
		return nil
	}
	return r.AbstractRules
}

func (r *RepeatedPatternRule) IsPremium() bool {
	if r == nil {
		return false
	}
	for _, pr := range r.PatternRules {
		if pr != nil && pr.IsPremium() {
			return true
		}
	}
	for _, ar := range r.AbstractRules {
		if ar != nil && ar.IsPremium() {
			return true
		}
	}
	return false
}

func (r *RepeatedPatternRule) defaultMaxDist() int {
	if r != nil && r.DefaultMaxDistanceTokens > 0 {
		return r.DefaultMaxDistanceTokens
	}
	return 60
}

// MatchSentences ports RepeatedPatternRule.match(List<AnalyzedSentence>).
// Document-relative LocalMatch offsets; only reports after min_prev_matches prior hits
// within distance_tokens (or defaultMaxDistanceTokens * min_prev_matches).
func (r *RepeatedPatternRule) MatchSentences(sentences []*languagetool.AnalyzedSentence) []languagetool.LocalMatch {
	if r == nil || len(sentences) == 0 {
		return nil
	}
	matchers := r.PatternRules
	if len(matchers) == 0 {
		// Fall back: build ephemeral PatternRules from abstracts (token sequences only).
		for _, ar := range r.AbstractRules {
			if ar == nil || len(ar.PatternTokens) == 0 {
				continue
			}
			pr := NewPatternRule(ar.ID, ar.LanguageCode, ar.PatternTokens, ar.Description, ar.Message, ar.ShortMessage)
			pr.MinPrevMatches = ar.MinPrevMatches
			pr.DistanceTokens = ar.DistanceTokens
			pr.Premium = ar.Premium
			pr.Tags = append([]rules.Tag(nil), ar.Tags...)
			matchers = append(matchers, pr)
		}
	}
	if len(matchers) == 0 {
		return nil
	}

	var out []languagetool.LocalMatch
	offsetChars := 0
	offsetTokens := 0
	prevFromToken := 0
	prevMatches := 0
	var distancesBetweenMatches []int
	filter := rules.NewSameRuleGroupFilter()

	for _, s := range sentences {
		if s == nil {
			continue
		}
		var sentenceMatches []*rules.RuleMatch
		for _, pr := range matchers {
			if pr == nil {
				continue
			}
			ms, err := pr.Match(s)
			if err != nil || len(ms) == 0 {
				continue
			}
			sentenceMatches = append(sentenceMatches, ms...)
		}
		sentenceMatches = filter.Filter(sentenceMatches)
		toks := s.GetTokensWithoutWhitespace()
		sentenceLenTokens := len(toks)
		for _, m := range sentenceMatches {
			if m == nil {
				continue
			}
			fromToken := 0
			for fromToken < sentenceLenTokens && toks[fromToken] != nil && toks[fromToken].GetStartPos() < m.GetFromPos() {
				fromToken++
			}
			fromToken += offsetTokens
			fromPos := m.GetFromPos() + offsetChars
			toPos := m.GetToPos() + offsetChars

			minPrev := 0
			distTok := 0
			if g, ok := m.Rule.(interface{ GetMinPrevMatches() int }); ok {
				minPrev = g.GetMinPrevMatches()
			}
			if g, ok := m.Rule.(interface{ GetDistanceTokens() int }); ok {
				distTok = g.GetDistanceTokens()
			}
			// Prefer primary rule metadata when Rule interface lacks getters.
			if minPrev == 0 && len(matchers) > 0 {
				minPrev = matchers[0].GetMinPrevMatches()
			}
			if distTok == 0 && len(matchers) > 0 {
				distTok = matchers[0].GetDistanceTokens()
			}
			maxDistanceTokens := distTok
			if maxDistanceTokens < 1 {
				maxDistanceTokens = r.defaultMaxDist() * minPrev
			}
			distancesBetweenMatches = append(distancesBetweenMatches, fromToken-prevFromToken)
			if prevMatches >= minPrev && isDistanceValid(distancesBetweenMatches, maxDistanceTokens, minPrev) {
				lm := rules.ToLocalMatches([]*rules.RuleMatch{m})
				if len(lm) > 0 {
					lm[0].FromPos = fromPos
					lm[0].ToPos = toPos
					if lm[0].RuleID == "" {
						lm[0].RuleID = r.GetID()
					}
					out = append(out, lm[0])
				}
			}
			prevFromToken = fromToken
			prevMatches++
		}
		offsetChars += len(s.GetText())
		// Java: -1 → not counting SENT_START
		if sentenceLenTokens > 0 {
			offsetTokens += sentenceLenTokens - 1
		}
	}
	return out
}

// isDistanceValid ports RepeatedPatternRule.isDistanceValid.
func isDistanceValid(distancesBetweenMatches []int, maxDistanceTokens, minPrevMatches int) bool {
	size := len(distancesBetweenMatches)
	distance := 0
	i := 0
	for i < minPrevMatches && i < size {
		distance += distancesBetweenMatches[size-1-i]
		i++
	}
	return distance < maxDistanceTokens
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
