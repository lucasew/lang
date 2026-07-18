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
	TokenException             string
	TokenExceptionRE           bool
	TokenExceptionCaseSensitive bool // when true, exception compares with exact case (LT case_sensitive on <exception>)
	// PreviousException ports Java exception scope="previous" (soft surface/regexp).
	PreviousException             string
	PreviousExceptionRE           bool
	PreviousExceptionCaseSensitive bool
	// NextException ports Java exception scope="next" (soft surface/regexp).
	NextException             string
	NextExceptionRE           bool
	NextExceptionCaseSensitive bool
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

func (p *PatternToken) SetStringPosException(tokenException string, regexp bool) {
	p.SetStringPosExceptionCS(tokenException, regexp, false)
}

// SetStringPosExceptionCS sets a surface exception with optional regexp and case sensitivity.
func (p *PatternToken) SetStringPosExceptionCS(tokenException string, regexp, caseSensitive bool) {
	p.TokenException = tokenException
	p.TokenExceptionRE = regexp
	p.TokenExceptionCaseSensitive = caseSensitive
}

// SetPreviousException ports Java exception scope="previous" (soft surface/regexp).
func (p *PatternToken) SetPreviousException(tokenException string, regexp, caseSensitive bool) {
	p.PreviousException = tokenException
	p.PreviousExceptionRE = regexp
	p.PreviousExceptionCaseSensitive = caseSensitive
}

// HasPreviousException reports whether a scope=previous exception is set.
func (p *PatternToken) HasPreviousException() bool {
	return p != nil && p.PreviousException != ""
}

// SetNextException ports Java exception scope="next" (soft surface/regexp).
func (p *PatternToken) SetNextException(tokenException string, regexp, caseSensitive bool) {
	p.NextException = tokenException
	p.NextExceptionRE = regexp
	p.NextExceptionCaseSensitive = caseSensitive
}

// HasNextException reports whether a scope=next exception is set.
func (p *PatternToken) HasNextException() bool {
	return p != nil && p.NextException != ""
}

// IsMatched ports PatternToken.isMatched for a single AnalyzedToken reading.
func (p *PatternToken) IsMatched(token *languagetool.AnalyzedToken) bool {
	return NewPatternTokenMatcher(p).IsMatched(token)
}
