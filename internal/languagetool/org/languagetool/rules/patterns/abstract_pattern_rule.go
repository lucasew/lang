package patterns

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// AbstractPatternRule ports shared fields of org.languagetool.rules.patterns.AbstractPatternRule.
type AbstractPatternRule struct {
	ID                      string
	SubID                   string
	Description             string
	LanguageCode            string
	PatternTokens           []*PatternToken
	Message                 string
	ShortMessage            string
	// CategoryID / CategoryName from surrounding <category> in soft grammar XML.
	CategoryID              string
	CategoryName            string
	// CategoryDefaultOff ports Category.isDefaultOff (XML category default="off").
	// Rules stay default-on; Check uses ignoreRule category branch (Java JLanguageTool.ignoreRule).
	CategoryDefaultOff bool
	// CategoryType ports category type="…" for ITS when the rule has no type (Java categoryIssueType).
	CategoryType string
	SuggestionsOutMsg       string
	SourceFile              string
	Filter                  RuleFilter
	FilterArgs              string
	StartPositionCorrection int
	EndPositionCorrection   int
	TestUnification         bool
	GetUnified              bool
	SentStart               bool
	AdjustSuggestionCase    bool
	SuggestionMatches       []*Match
	SuggestionMatchesOutMsg []*Match
	AntiPatterns            []*PatternRule // lightweight stand-in for DisambiguationPatternRule
	// UnifierConfig is the language-level <unification> table from the same grammar file.
	UnifierConfig *UnifierConfiguration
	// InterpretPreDisambig ports pattern raw_pos="yes".
	InterpretPreDisambig bool
	LineNumber           int
	DistanceTokens       int
	Premium              bool
	// DefaultOff is true when XML default="off" or default="temp_off" (Java Rule.defaultOff).
	DefaultOff bool
	// DefaultTempOff is true when XML default="temp_off" (Java Rule.defaultTempOff;
	// still DefaultOff; re-enabled by enableTempOff / enableTempOffRules).
	DefaultTempOff bool
	// ToneTags ports Rule.toneTags (XML tone_tags on rule/rulegroup/category).
	ToneTags []languagetool.ToneTag
	// GoalSpecific ports Rule.isGoalSpecific (XML is_goal_specific).
	GoalSpecific bool
	// Tags ports Rule.tags (XML tags e.g. picky) for hasTag / LocalMatch.IsPicky.
	Tags []rules.Tag
}

func NewAbstractPatternRule(id, description, languageCode string, patternTokens []*PatternToken, getUnified bool) *AbstractPatternRule {
	return &AbstractPatternRule{
		ID:                   id,
		Description:          description,
		LanguageCode:         languageCode,
		PatternTokens:        append([]*PatternToken(nil), patternTokens...),
		GetUnified:           getUnified,
		AdjustSuggestionCase: true,
		LineNumber:           -1,
		SubID:                "1",
	}
}

func (r *AbstractPatternRule) GetID() string          { return r.ID }
func (r *AbstractPatternRule) GetSubId() string       { return r.SubID }
func (r *AbstractPatternRule) GetDescription() string { return r.Description }
func (r *AbstractPatternRule) GetFullId() string {
	if r.SubID == "" {
		return r.ID
	}
	return r.ID + "[" + r.SubID + "]"
}
func (r *AbstractPatternRule) GetMessage() string      { return r.Message }
func (r *AbstractPatternRule) GetShortMessage() string { return r.ShortMessage }
func (r *AbstractPatternRule) GetPatternTokens() []*PatternToken {
	return r.PatternTokens
}
func (r *AbstractPatternRule) SetMessage(m string)      { r.Message = m }
func (r *AbstractPatternRule) SetShortMessage(m string) { r.ShortMessage = m }
func (r *AbstractPatternRule) SetFilter(f RuleFilter)   { r.Filter = f }
func (r *AbstractPatternRule) SetFilterArgs(a string)   { r.FilterArgs = a }
func (r *AbstractPatternRule) GetDistanceTokens() int   { return r.DistanceTokens }
func (r *AbstractPatternRule) IsPremium() bool          { return r.Premium }
func (r *AbstractPatternRule) SetPremium(v bool)        { r.Premium = v }

// GetToneTags ports Rule.getToneTags.
func (r *AbstractPatternRule) GetToneTags() []languagetool.ToneTag {
	if r == nil || len(r.ToneTags) == 0 {
		return nil
	}
	return append([]languagetool.ToneTag(nil), r.ToneTags...)
}

// IsGoalSpecific ports Rule.isGoalSpecific.
func (r *AbstractPatternRule) IsGoalSpecific() bool {
	return r != nil && r.GoalSpecific
}
