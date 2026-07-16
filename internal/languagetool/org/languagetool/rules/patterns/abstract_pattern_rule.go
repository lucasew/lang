package patterns

// AbstractPatternRule ports shared fields of org.languagetool.rules.patterns.AbstractPatternRule.
type AbstractPatternRule struct {
	ID                      string
	SubID                   string
	Description             string
	LanguageCode            string
	PatternTokens           []*PatternToken
	Message                 string
	ShortMessage            string
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
	LineNumber              int
	DistanceTokens          int
	Premium                 bool
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
