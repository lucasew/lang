package patterns

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// PosToken ports PatternToken.PosToken.
type PosToken struct {
	PosTag string
	Regexp bool
	Negate bool
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
	// TokenExceptionNegation / PosNegation port exception negate / negate_pos.
	TokenExceptionNegation    bool
	TokenExceptionPosNegation bool
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
	p.Pos = &pos
}

func (p *PatternToken) SetWhitespaceBefore(v bool) {
	p.WhitespaceBefore = &v
}

func (p *PatternToken) SetMinOccurrence(n int) { p.MinOccurrence = n }
func (p *PatternToken) SetMaxOccurrence(n int) { p.MaxOccurrence = n }
func (p *PatternToken) SetNegation(v bool)     { p.Negation = v }
func (p *PatternToken) SetSkipNext(n int)      { p.SkipNext = n }
func (p *PatternToken) SetInsideMarker(v bool) { p.InsideMarker = v }

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

func (p *PatternToken) SetStringPosException(tokenException string, regexp bool) {
	p.SetStringPosExceptionCS(tokenException, regexp, false)
}

// SetStringPosExceptionCS sets a surface exception with optional regexp and case sensitivity.
func (p *PatternToken) SetStringPosExceptionCS(tokenException string, regexp, caseSensitive bool) {
	p.TokenException = tokenException
	p.TokenExceptionRE = regexp
	p.TokenExceptionCaseSensitive = caseSensitive
}

// SetStringPosExceptionFull ports Java PatternToken.setStringPosException with optional POS.
// Surface and/or postag may be set (POS-only exceptions are valid Java).
func (p *PatternToken) SetStringPosExceptionFull(tokenException string, surfaceRE, caseSensitive bool, posTag string, posRE bool) {
	p.SetStringPosExceptionFullNeg(tokenException, surfaceRE, caseSensitive, false, posTag, posRE, false)
}

// SetStringPosExceptionFullNeg ports setStringPosException with surface/POS negation flags.
func (p *PatternToken) SetStringPosExceptionFullNeg(tokenException string, surfaceRE, caseSensitive, negation bool, posTag string, posRE, posNegation bool) {
	p.TokenException = tokenException
	p.TokenExceptionRE = surfaceRE
	p.TokenExceptionCaseSensitive = caseSensitive
	p.TokenExceptionNegation = negation
	p.TokenExceptionPosTag = posTag
	p.TokenExceptionPosRE = posRE
	p.TokenExceptionPosNegation = posNegation
}

// HasCurrentException reports whether a current-scope exception (surface and/or POS) is set.
func (p *PatternToken) HasCurrentException() bool {
	return p != nil && (p.TokenException != "" || p.TokenExceptionPosTag != "")
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
