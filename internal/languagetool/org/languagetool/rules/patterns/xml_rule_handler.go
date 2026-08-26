package patterns

// XML attribute/element names from org.languagetool.rules.patterns.XMLRuleHandler.
const (
	XMLAttrID   = "id"
	XMLAttrName = "name"

	XMLPremium        = "premium"
	XMLYes            = "yes"
	XMLNo             = "no"
	XMLOff            = "off"
	XMLOn             = "on"
	XMLTempOff        = "temp_off"
	XMLTrue           = "true"
	XMLFalse          = "false"
	XMLGoalSpecific   = "is_goal_specific"
	XMLPostag         = "postag"
	XMLChunkTag       = "chunk"
	XMLChunkTagRE     = "chunk_re"
	XMLPostagRegexp   = "postag_regexp"
	XMLRegexp         = "regexp"
	XMLNegate         = "negate"
	XMLInflected      = "inflected"
	XMLNegatePos      = "negate_pos"
	XMLMarker         = "marker"
	XMLDefault        = "default"
	XMLType           = "type"
	XMLSpaceBefore    = "spacebefore"
	XMLExample        = "example"
	XMLScope          = "scope"
	XMLIgnore         = "ignore"
	XMLSkip           = "skip"
	XMLMin            = "min"
	XMLMax            = "max"
	XMLToken          = "token"
	XMLFeature        = "feature"
	XMLUnify          = "unify"
	XMLUnifyIgnore    = "unify-ignore"
	XMLAnd            = "and"
	XMLOr             = "or"
	XMLException      = "exception"
	XMLCaseSensitive  = "case_sensitive"
	XMLMark           = "mark"
	XMLPattern        = "pattern"
	XMLAntipattern    = "antipattern"
	XMLMatch          = "match"
	XMLUnification    = "unification"
	XMLRule           = "rule"
	XMLRules          = "rules"
	XMLRuleGroup      = "rulegroup"
	XMLPhrases        = "phrases"
	XMLMessage        = "message"
	XMLSuggestion     = "suggestion"
	XMLTabName        = "tab"
	XMLMinPrevMatches = "min_prev_matches"
	XMLDistanceTokens = "distance_tokens"
	XMLPrio           = "prio"
)

// RegexpMode ports XMLRuleHandler.RegexpMode.
type RegexpMode string

const (
	RegexpModeSmart RegexpMode = "SMART"
	RegexpModeExact RegexpMode = "EXACT"
)

// XMLRuleHandler is a shared base for pattern / disambiguation XML loaders
// (ports field surface of org.languagetool.rules.patterns.XMLRuleHandler).
type XMLRuleHandler struct {
	Rules        []*AbstractPatternRule
	LanguageCode string
	// RelaxedMode skips missing id/name errors (online editor).
	RelaxedMode bool

	// Parse buffers (mirrors Java StringBuilders).
	CorrectExample   string
	IncorrectExample string
	Message          string
	SuggestionsOut   string
	Match            string
	ID               string
	Name             string
	SourceFile       string

	InRuleGroup bool
	InPattern   bool
	InToken     bool
	InException bool
	InMarker    bool

	PatternTokens []*PatternToken
	// Exceptions/and-groups deferred.
}

func NewXMLRuleHandler(languageCode string) *XMLRuleHandler {
	return &XMLRuleHandler{LanguageCode: languageCode}
}

func (h *XMLRuleHandler) SetRelaxedMode(v bool) { h.RelaxedMode = v }

func (h *XMLRuleHandler) GetRules() []*AbstractPatternRule { return h.Rules }

// ResetPattern clears pattern-local state between rules.
func (h *XMLRuleHandler) ResetPattern() {
	h.PatternTokens = nil
	h.InPattern = false
	h.InToken = false
	h.InException = false
	h.InMarker = false
	h.Message = ""
	h.SuggestionsOut = ""
	h.Match = ""
	h.CorrectExample = ""
	h.IncorrectExample = ""
}

// AttrYes reports whether attr is "yes"/"true"/"on".
func AttrYes(v string) bool {
	switch v {
	case XMLYes, XMLTrue, XMLOn:
		return true
	default:
		return false
	}
}

// FinishRule builds an AbstractPatternRule from current buffers and appends it.
func (h *XMLRuleHandler) FinishRule() *AbstractPatternRule {
	if h.ID == "" && !h.RelaxedMode {
		return nil
	}
	r := NewAbstractPatternRule(h.ID, h.Name, h.LanguageCode, h.PatternTokens, false)
	r.Message = h.Message
	r.SourceFile = h.SourceFile
	h.Rules = append(h.Rules, r)
	return r
}
