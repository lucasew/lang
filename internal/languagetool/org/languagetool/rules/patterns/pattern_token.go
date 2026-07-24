package patterns

import (
	"sort"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// NormalizeTextPattern ports PatternToken.normalizeTextPattern —
// null → "" ; StringTools.trimWhitespace.
func NormalizeTextPattern(token string) string {
	return tools.TrimWhitespace(token)
}

// UnknownPOSTag ports PatternToken.UNKNOWN_TAG ("UNKNOWN").
const UnknownPOSTag = "UNKNOWN"

// PosToken ports PatternToken.PosToken.
type PosToken struct {
	PosTag string
	Regexp bool
	Negate bool
	// posPattern ports PosToken.posPattern (StringMatcher for regexp POS).
	posPattern *StringMatcher
	// posUnknown ports PosToken.posUnknown — pattern accepts untagged tokens.
	posUnknown bool
	// prepared is true after ensurePosMatcher has run for current fields.
	prepared bool
}

// ensurePosMatcher ports PosToken ctor: StringMatcher.regexp + posUnknown flag.
func (p *PosToken) ensurePosMatcher() {
	if p == nil || p.prepared {
		return
	}
	p.prepared = true
	if p.Regexp {
		pat := normalizeJavaRegexp(p.PosTag)
		p.posPattern = NewStringMatcherRegexp(pat)
		p.posUnknown = p.posPattern.Matches(UnknownPOSTag)
		return
	}
	p.posPattern = nil
	p.posUnknown = p.PosTag == UnknownPOSTag
}

// IsPosTokenMatched ports PatternToken.isPosTokenMatched.
func IsPosTokenMatched(pos *PosToken, token *languagetool.AnalyzedToken) bool {
	if pos == nil || pos.PosTag == "" {
		// Java: pos == null || pos.posTag == null → true
		return true
	}
	pos.ensurePosMatcher()
	if pos.posUnknown && token != nil && token.HasNoTag() {
		return true
	}
	if token == nil {
		return false
	}
	tokenPos := token.GetPOSTag()
	if tokenPos == nil {
		return false
	}
	if pos.posPattern != nil {
		return pos.posPattern.Matches(*tokenPos)
	}
	return *tokenPos == pos.PosTag
}

// PatternToken ports a subset of org.languagetool.rules.patterns.PatternToken
// needed by PatternTokenBuilder and anti-pattern helpers.
type PatternToken struct {
	Token            string
	CaseSensitive    bool
	Regexp           bool
	MatchInflected   bool
	Pos              *PosToken
	Negation         bool
	MinOccurrence    int
	MaxOccurrence    int
	SkipNext         int
	InsideMarker     bool
	WhitespaceBefore *bool // nil = unset
	// ChunkTag ports Java PatternToken chunk / chunk_re (ChunkTag on readings).
	ChunkTag       string
	ChunkTagRegexp bool
	// AndGroup ports Java PatternToken and-group: other PatternTokens that must
	// each match *some* reading of the same token (not necessarily the same reading).
	AndGroup []*PatternToken
	// OrGroup ports Java PatternToken or-group: alternatives expanded at load time
	// (PatternRuleHandler.createRules). The base token is alternative 0; OrGroup
	// holds the remaining alternatives (not used at match time after expansion).
	OrGroup []*PatternToken
	TokenException             string
	TokenExceptionRE           bool
	TokenExceptionCaseSensitive bool // when true, exception compares with exact case (LT case_sensitive on <exception>)
	// TokenExceptionPos ports Java <exception postag="…"> (optional POS gate on
	// the same exception PatternToken). Empty surface + POS-only is valid Java.
	TokenExceptionPosTag string
	TokenExceptionPosRE  bool
	// TokenExceptionNegation / PosNegation port exception negate / negate_pos
	// for the first current-scope exception (legacy fields; prefer CurrentExceptions).
	TokenExceptionNegation    bool
	TokenExceptionPosNegation bool
	// CurrentExceptions ports non-next entries of rareFields.currentAndNextExceptions
	// (scope empty = current). Multi-exception disjunction via isExceptionMatched.
	CurrentExceptions []*PatternToken
	// PreviousExceptions ports Java rareFields.previousExceptions (scope="previous").
	// Each element is a full exception PatternToken (surface/POS/negation/inflected).
	PreviousExceptions []*PatternToken
	// NextExceptions ports next-scope entries of rareFields.currentAndNextExceptions
	// (exception PatternTokens with EXCEPTION_VALID_NEXT).
	NextExceptions []*PatternToken
	// PreviousException / NextException are convenience surfaces for the first
	// scope=previous/next exception (tests + simple builders). Prefer lists.
	PreviousException              string
	PreviousExceptionRE            bool
	PreviousExceptionCaseSensitive bool
	NextException              string
	NextExceptionRE            bool
	NextExceptionCaseSensitive bool
	// UniFeatures ports Java PatternToken unificationFeatures (feature → types).
	// Non-nil means the token participates in unification (isUnified).
	// Empty type list means all registered types for that feature (Java).
	UniFeatures map[string][]string
	// UniNegated ports isUniNegated — set on last token of <unify negate="yes">.
	UniNegated bool
	// LastInUnification ports isLastInUnification — last token in a <unify> block.
	LastInUnification bool
	// UnificationNeutral ports isUnificationNeutral — <unify-ignore> token.
	UnificationNeutral bool
	// TokenMatch ports PatternToken tokenReference (<match no="…" setpos="yes"/> inside token).
	// TokenRef is the raw XML no= value used as offset from firstMatchToken (Java resolveReference).
	TokenMatch *Match
	// PhraseName ports PatternToken.phraseName (set by preparePhrase / phraseref idref).
	// Non-empty means isPartOfPhrase — used for PatternRule.elementNo / useList.
	PhraseName string
}

func NewPatternToken(token string, caseSensitive, regexp, matchInflected bool) *PatternToken {
	return &PatternToken{
		Token:          token,
		CaseSensitive:  caseSensitive,
		Regexp:         regexp,
		MatchInflected: matchInflected,
		MinOccurrence:  1,
		MaxOccurrence:  1,
		InsideMarker:   true,
	}
}

func (p *PatternToken) SetPosToken(pos PosToken) {
	pos.prepared = false
	pos.posPattern = nil
	p.Pos = &pos
}

// IsSentenceStart ports PatternToken.isSentenceStart.
func (p *PatternToken) IsSentenceStart() bool {
	return p != nil && p.Pos != nil &&
		p.Pos.PosTag == languagetool.SentenceStartTagName && !p.Pos.Negate
}

func (p *PatternToken) SetWhitespaceBefore(v bool) {
	p.WhitespaceBefore = &v
}

func (p *PatternToken) SetMinOccurrence(n int) { p.MinOccurrence = n }
func (p *PatternToken) SetMaxOccurrence(n int) { p.MaxOccurrence = n }
func (p *PatternToken) SetNegation(v bool)     { p.Negation = v }
func (p *PatternToken) SetSkipNext(n int)      { p.SkipNext = n }
func (p *PatternToken) SetInsideMarker(v bool) { p.InsideMarker = v }

// SetPhraseName ports PatternToken.setPhraseName (phraseref idref).
func (p *PatternToken) SetPhraseName(id string) {
	if p != nil {
		p.PhraseName = id
	}
}

// GetPhraseName ports PatternToken.getPhraseName.
func (p *PatternToken) GetPhraseName() string {
	if p == nil {
		return ""
	}
	return p.PhraseName
}

// IsPartOfPhrase ports PatternToken.isPartOfPhrase (phraseName != null).
func (p *PatternToken) IsPartOfPhrase() bool {
	return p != nil && p.PhraseName != ""
}

// SetChunkTag ports PatternToken.setChunkTag (exact or regexp chunk name).
func (p *PatternToken) SetChunkTag(tag string, regexp bool) {
	p.ChunkTag = tag
	p.ChunkTagRegexp = regexp
}

// AddAndGroupElement ports PatternToken.setAndGroupElement.
func (p *PatternToken) AddAndGroupElement(andTok *PatternToken) {
	if p == nil || andTok == nil {
		return
	}
	p.AndGroup = append(p.AndGroup, andTok)
}

// AddOrGroupElement ports PatternToken.setOrGroupElement.
func (p *PatternToken) AddOrGroupElement(orTok *PatternToken) {
	if p == nil || orTok == nil {
		return
	}
	p.OrGroup = append(p.OrGroup, orTok)
}

// HasOrGroup ports PatternToken.hasOrGroup.
func (p *PatternToken) HasOrGroup() bool {
	return p != nil && len(p.OrGroup) > 0
}

// HasAndGroup ports PatternToken.hasAndGroup.
func (p *PatternToken) HasAndGroup() bool {
	return p != nil && len(p.AndGroup) > 0
}

// GetPOStag ports PatternToken.getPOStag.
func (p *PatternToken) GetPOStag() string {
	if p == nil || p.Pos == nil {
		return ""
	}
	return p.Pos.PosTag
}

// GetPOSNegation ports PatternToken.getPOSNegation.
func (p *PatternToken) GetPOSNegation() bool {
	return p != nil && p.Pos != nil && p.Pos.Negate
}

// IsInflected ports PatternToken.isInflected.
func (p *PatternToken) IsInflected() bool {
	return p != nil && p.MatchInflected
}

// HasCurrentOrNextExceptions ports PatternToken.hasCurrentOrNextExceptions.
func (p *PatternToken) HasCurrentOrNextExceptions() bool {
	if p == nil {
		return false
	}
	return p.HasCurrentException() || p.HasNextException()
}

// AddCurrentException ports PatternToken.addException(scopeNext=false, scopePrevious=false).
func (p *PatternToken) AddCurrentException(ex *PatternToken) {
	if p == nil || ex == nil {
		return
	}
	p.CurrentExceptions = append(p.CurrentExceptions, ex)
	// Mirror first exception into legacy TokenException* fields for tests/builders.
	if len(p.CurrentExceptions) == 1 {
		p.TokenException = ex.Token
		p.TokenExceptionRE = ex.Regexp
		p.TokenExceptionCaseSensitive = ex.CaseSensitive
		p.TokenExceptionNegation = ex.Negation
		if ex.Pos != nil {
			p.TokenExceptionPosTag = ex.Pos.PosTag
			p.TokenExceptionPosRE = ex.Pos.Regexp
			p.TokenExceptionPosNegation = ex.Pos.Negate
		}
	}
}

func (p *PatternToken) SetStringPosException(tokenException string, regexp bool) {
	p.SetStringPosExceptionCS(tokenException, regexp, false)
}

// SetStringPosExceptionCS sets a surface exception with optional regexp and case sensitivity.
func (p *PatternToken) SetStringPosExceptionCS(tokenException string, regexp, caseSensitive bool) {
	p.SetStringPosExceptionFullNeg(tokenException, regexp, caseSensitive, false, "", false, false)
}

// SetStringPosExceptionFull ports Java PatternToken.setStringPosException with optional POS.
// Surface and/or postag may be set (POS-only exceptions are valid Java).
func (p *PatternToken) SetStringPosExceptionFull(tokenException string, surfaceRE, caseSensitive bool, posTag string, posRE bool) {
	p.SetStringPosExceptionFullNeg(tokenException, surfaceRE, caseSensitive, false, posTag, posRE, false)
}

// SetStringPosExceptionFullNeg ports setStringPosException with surface/POS negation flags.
// Appends a current-scope exception PatternToken (Java always adds; multi is a disjunction).
func (p *PatternToken) SetStringPosExceptionFullNeg(tokenException string, surfaceRE, caseSensitive, negation bool, posTag string, posRE, posNegation bool) {
	if p == nil {
		return
	}
	ex := NewPatternToken(tokenException, caseSensitive, surfaceRE, false)
	ex.SetNegation(negation)
	if posTag != "" {
		ex.SetPosToken(PosToken{PosTag: posTag, Regexp: posRE, Negate: posNegation})
	}
	p.AddCurrentException(ex)
}

// HasCurrentException reports whether a current-scope exception (surface and/or POS) is set.
func (p *PatternToken) HasCurrentException() bool {
	if p == nil {
		return false
	}
	if len(p.CurrentExceptions) > 0 {
		return true
	}
	return p.TokenException != "" || p.TokenExceptionPosTag != ""
}

// AddPreviousException ports PatternToken.addException(..., scopePrevious=true).
func (p *PatternToken) AddPreviousException(ex *PatternToken) {
	if p == nil || ex == nil {
		return
	}
	p.PreviousExceptions = append(p.PreviousExceptions, ex)
	// Keep first-exception surface fields for simple getters/tests.
	if p.PreviousException == "" && ex.Token != "" {
		p.PreviousException = ex.Token
		p.PreviousExceptionRE = ex.Regexp
		p.PreviousExceptionCaseSensitive = ex.CaseSensitive
	}
}

// SetPreviousException ports a simple surface-only scope=previous exception.
// Prefer AddPreviousException for full POS/negation/multi support.
func (p *PatternToken) SetPreviousException(tokenException string, regexp, caseSensitive bool) {
	if p == nil {
		return
	}
	ex := NewPatternToken(tokenException, caseSensitive, regexp, false)
	p.AddPreviousException(ex)
}

// HasPreviousException ports PatternToken.hasPreviousException.
func (p *PatternToken) HasPreviousException() bool {
	if p == nil {
		return false
	}
	return len(p.PreviousExceptions) > 0 || p.PreviousException != ""
}

// AddNextException ports PatternToken.addException(scopeNext=true, ...).
func (p *PatternToken) AddNextException(ex *PatternToken) {
	if p == nil || ex == nil {
		return
	}
	p.NextExceptions = append(p.NextExceptions, ex)
	if p.NextException == "" && ex.Token != "" {
		p.NextException = ex.Token
		p.NextExceptionRE = ex.Regexp
		p.NextExceptionCaseSensitive = ex.CaseSensitive
	}
}

// SetNextException ports a simple surface-only scope=next exception.
func (p *PatternToken) SetNextException(tokenException string, regexp, caseSensitive bool) {
	if p == nil {
		return
	}
	ex := NewPatternToken(tokenException, caseSensitive, regexp, false)
	p.AddNextException(ex)
}

// HasNextException reports whether any scope=next exception is set.
func (p *PatternToken) HasNextException() bool {
	if p == nil {
		return false
	}
	return len(p.NextExceptions) > 0 || p.NextException != ""
}

// IsMatched ports PatternToken.isMatched for a single AnalyzedToken reading.
func (p *PatternToken) IsMatched(token *languagetool.AnalyzedToken) bool {
	return NewPatternTokenMatcher(p).IsMatched(token)
}

// IsUnified ports PatternToken.isUnified.
func (p *PatternToken) IsUnified() bool {
	return p != nil && p.UniFeatures != nil
}

// GetUniFeatures ports PatternToken.getUniFeatures.
func (p *PatternToken) GetUniFeatures() map[string][]string {
	if p == nil {
		return nil
	}
	return p.UniFeatures
}

// SetUnification ports PatternToken.setUnification (copies the feature map).
func (p *PatternToken) SetUnification(uniFeatures map[string][]string) {
	if p == nil {
		return
	}
	if uniFeatures == nil {
		p.UniFeatures = map[string][]string{}
		return
	}
	out := make(map[string][]string, len(uniFeatures))
	for k, v := range uniFeatures {
		out[k] = append([]string(nil), v...)
	}
	p.UniFeatures = out
}

// SetUniNegation ports PatternToken.setUniNegation.
func (p *PatternToken) SetUniNegation() {
	if p != nil {
		p.UniNegated = true
	}
}

// IsUniNegated ports PatternToken.isUniNegated.
func (p *PatternToken) IsUniNegated() bool {
	return p != nil && p.UniNegated
}

// SetLastInUnification ports PatternToken.setLastInUnification.
func (p *PatternToken) SetLastInUnification() {
	if p != nil {
		p.LastInUnification = true
	}
}

// IsLastInUnification ports PatternToken.isLastInUnification.
func (p *PatternToken) IsLastInUnification() bool {
	return p != nil && p.LastInUnification
}

// SetUnificationNeutral ports PatternToken.setUnificationNeutral.
func (p *PatternToken) SetUnificationNeutral() {
	if p != nil {
		p.UnificationNeutral = true
	}
}

// IsUnificationNeutral ports PatternToken.isUnificationNeutral.
func (p *PatternToken) IsUnificationNeutral() bool {
	return p != nil && p.UnificationNeutral
}

// IsReferenceElement ports PatternToken.isReferenceElement.
func (p *PatternToken) IsReferenceElement() bool {
	return p != nil && p.TokenMatch != nil
}

// SetMatch ports PatternToken.setMatch.
func (p *PatternToken) SetMatch(m *Match) {
	if p != nil {
		p.TokenMatch = m
	}
}

// GetMatch ports PatternToken.getMatch.
func (p *PatternToken) GetMatch() *Match {
	if p == nil {
		return nil
	}
	return p.TokenMatch
}

// CalcFormHints ports PatternToken.calcFormHints (performance; nil if unbounded).
func (p *PatternToken) CalcFormHints() []string {
	return p.calcStringHints(false)
}

// CalcLemmaHints ports PatternToken.calcLemmaHints.
func (p *PatternToken) CalcLemmaHints() []string {
	return p.calcStringHints(true)
}

// calcStringHints ports PatternToken.calcStringHints(inflected).
// Java returns null for empty sets after and/or retain/union.
func (p *PatternToken) calcStringHints(inflected bool) []string {
	if p == nil {
		return nil
	}
	if inflected != p.MatchInflected {
		return nil
	}
	result := p.calcOwnPossibleStringValues()
	if result == nil {
		return nil
	}
	if len(p.AndGroup) > 0 {
		set := stringSet(result)
		for _, t := range p.AndGroup {
			if t == nil {
				continue
			}
			h := t.calcStringHints(inflected)
			if h != nil {
				set = intersectStringSet(set, h)
			}
		}
		return stringSetSlice(set) // empty → nil
	}
	if len(p.OrGroup) > 0 {
		set := stringSet(result)
		for _, t := range p.OrGroup {
			if t == nil {
				continue
			}
			h := t.calcStringHints(inflected)
			if h == nil {
				return nil
			}
			for _, x := range h {
				set[x] = struct{}{}
			}
		}
		return stringSetSlice(set)
	}
	return result
}

// calcOwnPossibleStringValues ports PatternToken.calcOwnPossibleStringValues —
// textMatcher.getPossibleValues() (StringMatcher.create + getPossibleRegexpValues).
func (p *PatternToken) calcOwnPossibleStringValues() []string {
	if p == nil || p.Negation || !hasStringThatMustMatch(p) {
		return nil
	}
	// Java: return textMatcher.getPossibleValues();
	pat := NormalizeTextPattern(p.Token)
	if p.Regexp {
		pat = normalizeJavaRegexp(pat)
	}
	m := NewStringMatcher(pat, p.Regexp, p.CaseSensitive)
	vals := m.GetPossibleValues()
	if vals == nil || len(vals) == 0 {
		return nil
	}
	out := make([]string, 0, len(vals))
	for v := range vals {
		out = append(out, v)
	}
	// Deterministic order (Java HashSet is unordered).
	sort.Strings(out)
	return out
}

func stringSet(vals []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, v := range vals {
		m[v] = struct{}{}
	}
	return m
}

func intersectStringSet(a map[string]struct{}, b []string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, v := range b {
		if _, ok := a[v]; ok {
			out[v] = struct{}{}
		}
	}
	return out
}

func stringSetSlice(m map[string]struct{}) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	return out
}
