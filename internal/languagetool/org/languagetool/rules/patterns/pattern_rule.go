package patterns

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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
	// AntiPatterns ports Java AbstractPatternRule anti-patterns (IMMUNIZE via
	// DisambiguationPatternRule; Match re-runs on immunized sentence).
	AntiPatterns []*PatternRule
	// Filter / FilterArgs port AbstractPatternRule filter applied after pattern match.
	Filter     RuleFilter
	FilterArgs string
	// UnifierConfig ports Language.getUnifierConfiguration for testUnification.
	UnifierConfig *UnifierConfiguration
	// SuggestionMatches ports AbstractPatternRule.suggestionMatches for formatMatches.
	SuggestionMatches []*Match
	// SuggestionMatchesOutMsg ports AbstractPatternRule.suggestionMatchesOutMsg.
	SuggestionMatchesOutMsg []*Match
	// SuggestionsOutMsg ports AbstractPatternRule.suggestionsOutMsg.
	SuggestionsOutMsg string
	// SuggestionTemplates are <suggestion> bodies after ProcessRuleMessage (may contain \N).
	SuggestionTemplates []string
	// StartPositionCorrection ports AbstractPatternRule.startPositionCorrection.
	StartPositionCorrection int
	// EndPositionCorrection ports AbstractPatternRule.endPositionCorrection.
	EndPositionCorrection int
	// AdjustSuggestionCase ports AbstractPatternRule.adjustSuggestionCase (default true).
	AdjustSuggestionCase *bool // nil = true (Java default)
	// InterpretPreDisambig ports PatternRule.interpretPosTagsPreDisambiguation (raw_pos="yes").
	InterpretPreDisambig bool
	// ElementNo ports PatternRule.elementNo — token counts per XML-level element
	// when phrases are present (phrases count as one element spanning N tokens).
	ElementNo []int
	// UseList ports PatternRule.useList — true when any token is part of a phrase.
	// Enables translateElementNo / phraseLen for skip and suggestion synthesis.
	UseList bool
	// IsMemberOfDisjunctiveSet ports PatternRule.isMemberOfDisjunctiveSet
	// (OR on phraserefs in includephrases).
	IsMemberOfDisjunctiveSet bool
	// ToneTags ports Rule.toneTags from XML tone_tags attributes.
	ToneTags []languagetool.ToneTag
	// GoalSpecific ports Rule.isGoalSpecific from XML is_goal_specific.
	GoalSpecific bool
	// DefaultOff ports Rule.isDefaultOff (XML default="off" / "temp_off").
	DefaultOff bool
	// DefaultTempOff ports Rule.isDefaultTempOff (XML default="temp_off" only).
	DefaultTempOff bool
	// SubID ports AbstractPatternRule.subId (rulegroup child index as string).
	SubID string
	// SourceFile ports Rule.sourceFile (grammar XML path).
	SourceFile string
	// IssueType ports Rule.locQualityIssueType string form (e.g. "grammar", "misspelling").
	IssueType string
	// URL ports Rule.url.
	URL string
	// Priority ports Rule.priority (XML prio= inheritance).
	Priority int
	// Premium ports Rule.isPremium (XML premium= inheritance).
	Premium bool
	// MinPrevMatches ports AbstractPatternRule.minPrevMatches (XML min_prev_matches).
	MinPrevMatches int
	// DistanceTokens ports AbstractPatternRule.distanceTokens (XML distance_tokens).
	DistanceTokens int
}

func NewPatternRule(id, languageCode string, tokens []*PatternToken, description, message, shortMessage string) *PatternRule {
	pr := &PatternRule{
		ID:           id,
		LanguageCode: languageCode,
		Tokens:       append([]*PatternToken(nil), tokens...),
		Description:  description,
		Message:      message,
		ShortMessage: shortMessage,
	}
	pr.computeElementNo()
	return pr
}

// computeElementNo ports PatternRule constructor elementNo / useList calculation.
// Phrases count as single XML elements spanning multiple pattern tokens.
// Control flow matches Java bug-for-bug (including cnt not cleared after non-phrase).
func (r *PatternRule) computeElementNo() {
	if r == nil {
		return
	}
	r.ElementNo = nil
	r.UseList = false
	prevName := ""
	cnt := 0
	loopCnt := 0
	tempUseList := false
	for _, pToken := range r.Tokens {
		if pToken != nil && pToken.IsPartOfPhrase() {
			curName := pToken.GetPhraseName()
			if tools.IsEmptyStr(prevName) || prevName == curName {
				cnt++
				tempUseList = true
			} else {
				r.ElementNo = append(r.ElementNo, cnt)
				curName = ""
				cnt = 0
			}
			prevName = curName
			loopCnt++
			if loopCnt == len(r.Tokens) && !tools.IsEmptyStr(prevName) {
				r.ElementNo = append(r.ElementNo, cnt)
			}
		} else {
			if cnt > 0 {
				r.ElementNo = append(r.ElementNo, cnt)
			}
			r.ElementNo = append(r.ElementNo, 1)
			loopCnt++
		}
	}
	r.UseList = tempUseList
}

// GetElementNo ports PatternRule.getElementNo.
func (r *PatternRule) GetElementNo() []int {
	if r == nil {
		return nil
	}
	return r.ElementNo
}

// IsWithComplexPhrase ports PatternRule.isWithComplexPhrase.
func (r *PatternRule) IsWithComplexPhrase() bool {
	return r != nil && r.IsMemberOfDisjunctiveSet
}

// NotComplexPhrase ports PatternRule.notComplexPhrase.
func (r *PatternRule) NotComplexPhrase() {
	if r != nil {
		r.IsMemberOfDisjunctiveSet = false
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

// GetToneTags ports Rule.getToneTags.
func (r *PatternRule) GetToneTags() []languagetool.ToneTag {
	if r == nil || len(r.ToneTags) == 0 {
		return nil
	}
	return append([]languagetool.ToneTag(nil), r.ToneTags...)
}

// IsGoalSpecific ports Rule.isGoalSpecific.
func (r *PatternRule) IsGoalSpecific() bool {
	return r != nil && r.GoalSpecific
}

// IsDefaultOff ports Rule.isDefaultOff.
func (r *PatternRule) IsDefaultOff() bool {
	return r != nil && r.DefaultOff
}

// IsDefaultTempOff ports Rule.isDefaultTempOff.
func (r *PatternRule) IsDefaultTempOff() bool {
	return r != nil && r.DefaultTempOff
}

// GetSubID ports AbstractPatternRule.getSubId (empty when unset).
func (r *PatternRule) GetSubID() string {
	if r == nil {
		return ""
	}
	return r.SubID
}

// GetSourceFile ports Rule.getSourceFile.
func (r *PatternRule) GetSourceFile() string {
	if r == nil {
		return ""
	}
	return r.SourceFile
}

// GetURL ports Rule.getUrl.
func (r *PatternRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType as string.
func (r *PatternRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ""
	}
	return rules.ITSIssueType(r.IssueType)
}

// GetPriority ports Rule.getPriority.
func (r *PatternRule) GetPriority() int {
	if r == nil {
		return 0
	}
	return r.Priority
}

// IsPremium ports Rule.isPremium.
func (r *PatternRule) IsPremium() bool {
	return r != nil && r.Premium
}

// SetPremium ports Rule.setPremium.
func (r *PatternRule) SetPremium(v bool) {
	if r != nil {
		r.Premium = v
	}
}

// GetMinPrevMatches ports AbstractPatternRule.getMinPrevMatches.
func (r *PatternRule) GetMinPrevMatches() int {
	if r == nil {
		return 0
	}
	return r.MinPrevMatches
}

// GetDistanceTokens ports AbstractPatternRule.getDistanceTokens.
func (r *PatternRule) GetDistanceTokens() int {
	if r == nil {
		return 0
	}
	return r.DistanceTokens
}

// SetDefaultTempOff ports Rule.setDefaultTempOff (defaultOff + defaultTempOff).
func (r *PatternRule) SetDefaultTempOff() {
	if r == nil {
		return
	}
	r.DefaultOff = true
	r.DefaultTempOff = true
}

// HasToneTag ports Rule.hasToneTag.
func (r *PatternRule) HasToneTag(tag languagetool.ToneTag) bool {
	if r == nil {
		return false
	}
	for _, t := range r.ToneTags {
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

// Match ports PatternRule.match: canBeIgnoredFor fast-path, PatternRuleMatcher, antipatterns.
func (r *PatternRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	matcher := NewPatternRuleMatcherFromPattern(r)
	// Java PatternRule.match: if (canBeIgnoredFor(sentence)) return EMPTY_ARRAY
	if matcher.Rule != nil && matcher.Rule.CanBeIgnoredFor(sentence) {
		return nil, nil
	}
	found, err := matcher.Match(sentence)
	if err != nil {
		return found, err
	}
	return r.checkForAntiPatterns(sentence, matcher, found)
}

// checkForAntiPatterns ports PatternRule.checkForAntiPatterns:
// if any match and antipatterns exist, immunize via getSentenceWithImmunization
// and re-match when any token was immunized (Java PatternRuleMatcher skips them).
func (r *PatternRule) checkForAntiPatterns(
	sentence *languagetool.AnalyzedSentence,
	matcher *PatternRuleMatcher,
	matches []*rules.RuleMatch,
) ([]*rules.RuleMatch, error) {
	if r == nil || len(matches) == 0 || len(r.AntiPatterns) == 0 {
		return matches, nil
	}
	immunized := sentenceWithImmunization(sentence, r.AntiPatterns)
	if immunized == nil || !sentenceHasImmunizedToken(immunized) {
		return matches, nil
	}
	return matcher.Match(immunized)
}

// sentenceWithImmunization ports Rule.getSentenceWithImmunization for grammar
// <antipattern>s stored as PatternRule (Java: DisambiguationPatternRule IMMUNIZE).
func sentenceWithImmunization(sentence *languagetool.AnalyzedSentence, antis []*PatternRule) *languagetool.AnalyzedSentence {
	if sentence == nil || len(antis) == 0 {
		return sentence
	}
	immunized := sentence.Copy(sentence)
	if immunized == nil {
		return sentence
	}
	// Java AnalyzedSentence.copy reuses source getTokensWithoutWhitespace() refs.
	// Rebuild so nonBlank views the copy's tokens — IMMUNIZE then shows on getTokens()
	// (PatternRule.checkForAntiPatterns anyMatch isImmunized) and rematch SkipImmunized.
	immunized = languagetool.NewAnalyzedSentence(immunized.GetTokens())
	for _, ap := range antis {
		if ap == nil || len(ap.Tokens) == 0 {
			continue
		}
		// Java AbstractPatternRulePerformer (disambig) does not skip immunized tokens
		// while applying IMMUNIZE antipatterns.
		am := NewPatternRuleMatcherFromPattern(ap)
		am.SkipImmunized = false
		antiMatches, err := am.Match(immunized)
		if err != nil || len(antiMatches) == 0 {
			continue
		}
		nws := immunized.GetTokensWithoutWhitespace()
		for _, m := range antiMatches {
			if m == nil {
				continue
			}
			first, last := matchSpanTokenIndices(nws, m.FromPos, m.ToPos)
			if first < 0 || last < 0 {
				continue
			}
			for i := first; i <= last && i < len(nws); i++ {
				if nws[i] != nil {
					nws[i].Immunize(0)
				}
			}
		}
	}
	return immunized
}

func sentenceHasImmunizedToken(sentence *languagetool.AnalyzedSentence) bool {
	if sentence == nil {
		return false
	}
	for _, t := range sentence.GetTokens() {
		if t != nil && t.IsImmunized() {
			return true
		}
	}
	return false
}

// matchSpanTokenIndices maps a RuleMatch char span to non-whitespace token indices
// (same approach as DisambiguationPatternRule.Replace).
func matchSpanTokenIndices(nws []*languagetool.AnalyzedTokenReadings, fromPos, toPos int) (first, last int) {
	first, last = -1, -1
	for i, t := range nws {
		if t == nil {
			continue
		}
		if t.GetStartPos() == fromPos {
			first = i
		}
		if t.GetEndPos() == toPos || t.GetStartPos()+len(t.GetToken()) == toPos {
			last = i
		}
	}
	if first < 0 {
		for i, t := range nws {
			if t == nil {
				continue
			}
			if t.GetStartPos() >= fromPos && (first < 0 || t.GetStartPos() < nws[first].GetStartPos()) {
				first = i
			}
			if t.GetStartPos() < toPos {
				last = i
			}
		}
	}
	return first, last
}

// keepByGrammarAntiPatterns is retained for disambig-style overlap tests; PatternRule.Match
// uses immunization rematch (Java checkForAntiPatterns) instead.
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
