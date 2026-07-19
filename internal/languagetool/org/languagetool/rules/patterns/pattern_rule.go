package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// PatternRule ports a metadata surface of org.languagetool.rules.patterns.PatternRule
// (full matcher engine not yet wired).
type PatternRule struct {
	ID           string
	LanguageCode string // short code with optional variant
	Tokens       []*PatternToken
	Description  string
	Message      string
	ShortMessage string
	Tags         []rules.Tag
	// AntiPatterns ports Java AbstractPatternRule anti-patterns (suppress overlapping matches).
	AntiPatterns []*PatternRule
	// Filter / FilterArgs port AbstractPatternRule filter applied after pattern match.
	Filter     RuleFilter
	FilterArgs string
	// UnifierConfig ports Language.getUnifierConfiguration for testUnification.
	UnifierConfig *UnifierConfiguration
	// SuggestionMatches ports AbstractPatternRule.suggestionMatches for formatMatches.
	SuggestionMatches []*Match
	// SuggestionTemplates are <suggestion> bodies after ProcessRuleMessage (may contain \N).
	SuggestionTemplates []string
	// InterpretPreDisambig ports PatternRule.interpretPosTagsPreDisambiguation (raw_pos="yes").
	InterpretPreDisambig bool
}

func NewPatternRule(id, languageCode string, tokens []*PatternToken, description, message, shortMessage string) *PatternRule {
	return &PatternRule{
		ID:           id,
		LanguageCode: languageCode,
		Tokens:       append([]*PatternToken(nil), tokens...),
		Description:  description,
		Message:      message,
		ShortMessage: shortMessage,
	}
}

func (r *PatternRule) GetID() string          { return r.ID }
func (r *PatternRule) GetDescription() string { return r.Description }
func (r *PatternRule) GetMessage() string     { return r.Message }
func (r *PatternRule) GetTags() []rules.Tag   { return r.Tags }

// GetLanguageCode ports Language short code for AdaptSuggestionsFilter (Java getLanguage().getShortCode).
func (r *PatternRule) GetLanguageCode() string {
	if r == nil {
		return ""
	}
	return r.LanguageCode
}
func (r *PatternRule) SetTags(tags []rules.Tag) {
	r.Tags = append([]rules.Tag(nil), tags...)
}
func (r *PatternRule) HasTag(tag rules.Tag) bool {
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// SupportsLanguage reports whether the rule applies to the given short code (with optional variant).
// Empty LanguageCode only matches empty code (callers treat unset rules separately).
func (r *PatternRule) SupportsLanguage(code string) bool {
	if r == nil {
		return false
	}
	if r.LanguageCode == "" {
		return code == ""
	}
	if code == "" {
		return false
	}
	a, b := strings.ToLower(r.LanguageCode), strings.ToLower(code)
	if a == b {
		return true
	}
	// en matches en-US / en-GB and vice versa on base
	abase, bbase := a, b
	if i := strings.Index(a, "-"); i > 0 {
		abase = a[:i]
	}
	if i := strings.Index(b, "-"); i > 0 {
		bbase = b[:i]
	}
	return abase == bbase
}

// FalseFriendPatternRule ports org.languagetool.rules.patterns.FalseFriendPatternRule.
type FalseFriendPatternRule struct {
	*PatternRule
}

func NewFalseFriendPatternRule(id, languageCode string, tokens []*PatternToken, description, message, shortMessage string) *FalseFriendPatternRule {
	pr := NewPatternRule(id, languageCode, tokens, description, message, shortMessage)
	pr.SetTags([]rules.Tag{rules.TagPicky})
	return &FalseFriendPatternRule{PatternRule: pr}
}

// Match runs a simplified PatternRuleMatcher against the sentence.
// Java: matches suppressed when an antipattern overlaps (AbstractPatternRulePerformer).
func (r *PatternRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	found, err := NewPatternRuleMatcherFromPattern(r).Match(sentence)
	if err != nil || len(found) == 0 || len(r.AntiPatterns) == 0 {
		return found, err
	}
	var kept []*rules.RuleMatch
	for _, rm := range found {
		if rm == nil {
			continue
		}
		if keepByGrammarAntiPatterns(r.AntiPatterns, sentence, rm.FromPos, rm.ToPos) {
			kept = append(kept, rm)
		}
	}
	return kept, nil
}

// keepByGrammarAntiPatterns returns false when any antipattern match overlaps [from,to].
// Same overlap test as DisambiguationPatternRule.keepByDisambig / Java PatternRuleMatcher.
func keepByGrammarAntiPatterns(antis []*PatternRule, sentence *languagetool.AnalyzedSentence, fromPos, toPos int) bool {
	for _, ap := range antis {
		if ap == nil || len(ap.Tokens) == 0 {
			continue
		}
		antiMatches, err := NewPatternRuleMatcherFromPattern(ap).Match(sentence)
		if err != nil || len(antiMatches) == 0 {
			continue
		}
		for _, dm := range antiMatches {
			if dm == nil {
				continue
			}
			if (dm.FromPos <= fromPos && dm.ToPos >= fromPos) ||
				(dm.FromPos <= toPos && dm.ToPos >= toPos) ||
				(dm.FromPos >= fromPos && dm.ToPos <= toPos) {
				return false
			}
		}
	}
	return true
}
