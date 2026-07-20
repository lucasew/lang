package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConsistencyPatternRuleTransformer ports
// org.languagetool.rules.patterns.ConsistencyPatternRuleTransformer.
//
// Rule id convention (Java comment): PREFIX_GROUPOFRULES_FEATURE
// where PREFIX is Language.getConsistencyRulePrefix() (default
// PREFIXFORCONSISTENCYRULES_).
type ConsistencyPatternRuleTransformer struct {
	LanguageCode string
	// Prefix ports Language.getConsistencyRulePrefix; empty → tools.ConsistencyRulePrefix.
	Prefix string
	// AdjustMatch ports Language.adjustMatch(RuleMatch, List<String>); nil = identity.
	AdjustMatch func(m *rules.RuleMatch, features []string) *rules.RuleMatch
}

func NewConsistencyPatternRuleTransformer(languageCode string) *ConsistencyPatternRuleTransformer {
	return &ConsistencyPatternRuleTransformer{
		LanguageCode: languageCode,
		Prefix:       tools.ConsistencyRulePrefix,
	}
}

func (t *ConsistencyPatternRuleTransformer) prefix() string {
	if t != nil && t.Prefix != "" {
		return t.Prefix
	}
	return tools.ConsistencyRulePrefix
}

// GetMainRuleId ports ConsistencyPatternRuleTransformer.getMainRuleId:
// parts[0] + "_" + parts[1] after split on '_'.
func GetMainRuleId(id string) string {
	parts := strings.Split(id, "_")
	if len(parts) < 2 {
		return id
	}
	return parts[0] + "_" + parts[1]
}

// GetFeature ports ConsistencyPatternRuleTransformer.getFeature: parts[2].
func GetFeature(id string) string {
	parts := strings.Split(id, "_")
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}

// Transform ports apply(): only rules whose id starts with the consistency
// prefix become ConsistencyPatternRule groups (keyed by main rule id).
func (t *ConsistencyPatternRuleTransformer) Transform(rules []*AbstractPatternRule) TransformedRules {
	if len(rules) == 0 {
		return NewTransformedRules(nil, nil)
	}
	pfx := t.prefix()
	toTransform := map[string][]*AbstractPatternRule{}
	var order []string
	var remaining []*AbstractPatternRule
	for _, r := range rules {
		if r == nil {
			continue
		}
		if strings.HasPrefix(r.ID, pfx) {
			main := GetMainRuleId(r.ID)
			if _, ok := toTransform[main]; !ok {
				order = append(order, main)
			}
			toTransform[main] = append(toTransform[main], r)
			continue
		}
		remaining = append(remaining, r)
	}
	var transformed []any
	for _, main := range order {
		group := toTransform[main]
		transformed = append(transformed, &ConsistencyPatternRule{
			MainID:       main,
			LanguageCode: t.LanguageCode,
			AbstractRules: group,
			AdjustMatch:  t.AdjustMatch,
		})
	}
	return NewTransformedRules(remaining, transformed)
}

// ConsistencyPatternRule ports
// ConsistencyPatternRuleTransformer.ConsistencyPatternRule (TextLevelRule).
type ConsistencyPatternRule struct {
	MainID       string
	LanguageCode string
	// AbstractRules are the wrapped AbstractPatternRule twins (metadata).
	AbstractRules []*AbstractPatternRule
	// PatternRules are concrete matchers (preferred when set by RegisterGrammar).
	PatternRules []*PatternRule
	// AdjustMatch ports Language.adjustMatch; nil = identity.
	AdjustMatch func(m *rules.RuleMatch, features []string) *rules.RuleMatch
}

// ConsistencyPatternRuleGroup is a historical alias for ConsistencyPatternRule.
type ConsistencyPatternRuleGroup = ConsistencyPatternRule

func (r *ConsistencyPatternRule) GetID() string {
	if r == nil {
		return ""
	}
	if r.MainID != "" {
		return r.MainID
	}
	if len(r.PatternRules) > 0 && r.PatternRules[0] != nil {
		return GetMainRuleId(r.PatternRules[0].GetID())
	}
	if len(r.AbstractRules) > 0 && r.AbstractRules[0] != nil {
		return GetMainRuleId(r.AbstractRules[0].ID)
	}
	return ""
}

func (r *ConsistencyPatternRule) GetDescription() string {
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

func (r *ConsistencyPatternRule) GetWrappedRules() []*AbstractPatternRule {
	if r == nil {
		return nil
	}
	return r.AbstractRules
}

func (r *ConsistencyPatternRule) IsPremium() bool {
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

// MinToCheckParagraph ports ConsistencyPatternRule.minToCheckParagraph (Java returns 0).
func (r *ConsistencyPatternRule) MinToCheckParagraph() int { return 0 }

// MatchSentences ports ConsistencyPatternRule.match(List<AnalyzedSentence>).
// Collects sentence pattern hits across the document, counts feature suffixes,
// and reports matches for minority (or all-on-tie) features via adjustMatch.
func (r *ConsistencyPatternRule) MatchSentences(sentences []*languagetool.AnalyzedSentence) []languagetool.LocalMatch {
	if r == nil || len(sentences) == 0 {
		return nil
	}
	matchers := r.PatternRules
	if len(matchers) == 0 {
		for _, ar := range r.AbstractRules {
			if ar == nil || len(ar.PatternTokens) == 0 {
				continue
			}
			pr := NewPatternRule(ar.ID, ar.LanguageCode, ar.PatternTokens, ar.Description, ar.Message, ar.ShortMessage)
			pr.Premium = ar.Premium
			pr.Tags = append([]rules.Tag(nil), ar.Tags...)
			matchers = append(matchers, pr)
		}
	}
	if len(matchers) == 0 {
		return nil
	}

	var matches []*rules.RuleMatch
	offsetChars := 0
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
		for _, rm := range sentenceMatches {
			if rm == nil {
				continue
			}
			// Java: setSentencePosition then setOffsetPosition to document coords.
			rm.FromPosSentence = rm.GetFromPos()
			rm.ToPosSentence = rm.GetToPos()
			rm.FromPos = rm.GetFromPos() + offsetChars
			rm.ToPos = rm.GetToPos() + offsetChars
			matches = append(matches, rm)
		}
		offsetChars += len(s.GetText())
	}

	countFeatures := map[string]int{}
	for _, rm := range matches {
		feat := GetFeature(ruleIDOfMatch(rm))
		countFeatures[feat] = countFeatures[feat] + 1
	}
	if len(countFeatures) < 2 {
		// no inconsistency
		return nil
	}
	max := 0
	for _, c := range countFeatures {
		if c > max {
			max = c
		}
	}
	var featuresWithMax []string
	var featuresToKeep []string
	for feat, c := range countFeatures {
		if c == max {
			featuresWithMax = append(featuresWithMax, feat)
		} else {
			featuresToKeep = append(featuresToKeep, feat)
		}
	}
	featuresToSuggest := append([]string(nil), featuresWithMax...)
	if len(featuresWithMax) > 1 {
		// tie at max → report all features
		featuresToKeep = append(featuresToKeep, featuresWithMax...)
	}
	keepSet := map[string]bool{}
	for _, f := range featuresToKeep {
		keepSet[f] = true
	}

	var out []languagetool.LocalMatch
	for _, rm := range matches {
		if rm == nil {
			continue
		}
		if !keepSet[GetFeature(ruleIDOfMatch(rm))] {
			continue
		}
		adj := rm
		if r.AdjustMatch != nil {
			if m2 := r.AdjustMatch(rm, featuresToSuggest); m2 != nil {
				adj = m2
			}
		}
		lm := rules.ToLocalMatches([]*rules.RuleMatch{adj})
		if len(lm) > 0 {
			out = append(out, lm[0])
		}
	}
	return out
}

func ruleIDOfMatch(m *rules.RuleMatch) string {
	if m == nil {
		return ""
	}
	// Prefer SpecificRuleId when set (Java getSpecificRuleId path not used here).
	if m.SpecificRuleId != "" {
		return m.SpecificRuleId
	}
	if g, ok := m.Rule.(interface{ GetID() string }); ok {
		return g.GetID()
	}
	return ""
}
