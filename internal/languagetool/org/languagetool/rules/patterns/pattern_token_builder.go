package patterns

// PatternTokenBuilder ports org.languagetool.rules.patterns.PatternTokenBuilder.
type PatternTokenBuilder struct {
	token              string
	posTag             string
	marker             bool
	matchInflected     bool
	caseSensitive      bool
	regexp             bool
	negation           bool
	isWhiteSpaceSet    bool
	isWhiteSpaceBefore bool
	minOccurrence      int
	maxOccurrence      int
	skip               int
	tokenException     string
	hasToken           bool
	hasPos             bool
}

func NewPatternTokenBuilder() *PatternTokenBuilder {
	return &PatternTokenBuilder{
		marker:        true,
		minOccurrence: 1,
		maxOccurrence: 1,
	}
}

func (b *PatternTokenBuilder) Token(token string) *PatternTokenBuilder {
	b.token = token
	b.hasToken = true
	return b
}

func (b *PatternTokenBuilder) CsToken(token string) *PatternTokenBuilder {
	b.token = token
	b.caseSensitive = true
	b.hasToken = true
	return b
}

func (b *PatternTokenBuilder) TokenRegex(token string) *PatternTokenBuilder {
	b.token = token
	b.regexp = true
	b.hasToken = true
	return b
}

func (b *PatternTokenBuilder) CsTokenRegex(token string) *PatternTokenBuilder {
	b.token = token
	b.regexp = true
	b.caseSensitive = true
	b.hasToken = true
	return b
}

func (b *PatternTokenBuilder) Pos(posTag string) *PatternTokenBuilder {
	return b.pos(posTag, false, "")
}

func (b *PatternTokenBuilder) PosRegex(posTag string) *PatternTokenBuilder {
	return b.pos(posTag, true, "")
}

func (b *PatternTokenBuilder) PosRegexWithStringException(posTag, tokenExceptionRegex string) *PatternTokenBuilder {
	return b.pos(posTag, true, tokenExceptionRegex)
}

func (b *PatternTokenBuilder) pos(posTag string, regexp bool, tokenExceptionRegex string) *PatternTokenBuilder {
	b.posTag = posTag
	b.regexp = regexp
	b.hasPos = true
	if tokenExceptionRegex != "" {
		b.tokenException = tokenExceptionRegex
	}
	return b
}

func (b *PatternTokenBuilder) Min(val int) *PatternTokenBuilder {
	if val < 0 {
		panic("minOccurrence must be >= 0")
	}
	b.minOccurrence = val
	return b
}

func (b *PatternTokenBuilder) Max(val int) *PatternTokenBuilder {
	b.maxOccurrence = val
	return b
}

func (b *PatternTokenBuilder) Mark(isMarked bool) *PatternTokenBuilder {
	b.marker = isMarked
	return b
}

func (b *PatternTokenBuilder) Negate() *PatternTokenBuilder {
	b.negation = true
	return b
}

func (b *PatternTokenBuilder) SetSkip(skip int) *PatternTokenBuilder {
	b.skip = skip
	return b
}

func (b *PatternTokenBuilder) SetIsWhiteSpaceBefore(whiteSpaceBefore bool) *PatternTokenBuilder {
	b.isWhiteSpaceBefore = whiteSpaceBefore
	b.isWhiteSpaceSet = true
	return b
}

func (b *PatternTokenBuilder) MatchInflectedForms() *PatternTokenBuilder {
	b.matchInflected = true
	return b
}

func (b *PatternTokenBuilder) Build() *PatternToken {
	pt := NewPatternToken(b.token, b.caseSensitive, b.regexp, b.matchInflected)
	if b.hasPos {
		pt.SetPosToken(PosToken{PosTag: b.posTag, Regexp: b.regexp, Negate: false})
	}
	if b.isWhiteSpaceSet {
		pt.SetWhitespaceBefore(b.isWhiteSpaceBefore)
	}
	if b.maxOccurrence < b.minOccurrence {
		panic("minOccurrence must <= maxOccurrence")
	}
	if b.tokenException != "" {
		pt.SetStringPosException(b.tokenException, true)
	}
	pt.SetMinOccurrence(b.minOccurrence)
	pt.SetMaxOccurrence(b.maxOccurrence)
	pt.SetNegation(b.negation)
	pt.SetSkipNext(b.skip)
	pt.SetInsideMarker(b.marker)
	return pt
}
