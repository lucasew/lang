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
	TokenException   string
	TokenExceptionRE bool
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

func (p *PatternToken) SetStringPosException(tokenException string, regexp bool) {
	p.TokenException = tokenException
	p.TokenExceptionRE = regexp
}

// IsMatched ports PatternToken.isMatched for a single AnalyzedToken reading.
func (p *PatternToken) IsMatched(token *languagetool.AnalyzedToken) bool {
	return NewPatternTokenMatcher(p).IsMatched(token)
}
