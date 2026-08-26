package patterns

// PatternRuleBuilderHelper ports org.languagetool.rules.patterns.PatternRuleBuilderHelper.

func TokenRegex(s string) *PatternToken {
	return NewPatternTokenBuilder().TokenRegex(s).Build()
}

func PosRegex(s string) *PatternToken {
	return NewPatternTokenBuilder().PosRegex(s).Build()
}

func CsToken(s string) *PatternToken {
	return NewPatternTokenBuilder().CsToken(s).Build()
}

func Pos(s string) *PatternToken {
	return NewPatternTokenBuilder().Pos(s).Build()
}

func Token(s string) *PatternToken {
	return NewPatternTokenBuilder().Token(s).Build()
}

func Regex(regex string) *PatternToken {
	return NewPatternTokenBuilder().TokenRegex(regex).Build()
}

func CsRegex(regex string) *PatternToken {
	return NewPatternTokenBuilder().CsTokenRegex(regex).Build()
}

// PatternRuleBuilderHelper is the Java-name twin for pattern token builders.
type PatternRuleBuilderHelper struct{}

func (PatternRuleBuilderHelper) TokenRegex(s string) *PatternToken { return TokenRegex(s) }
func (PatternRuleBuilderHelper) PosRegex(s string) *PatternToken   { return PosRegex(s) }
func (PatternRuleBuilderHelper) CsToken(s string) *PatternToken    { return CsToken(s) }
func (PatternRuleBuilderHelper) Pos(s string) *PatternToken        { return Pos(s) }
func (PatternRuleBuilderHelper) Token(s string) *PatternToken      { return Token(s) }
func (PatternRuleBuilderHelper) Regex(s string) *PatternToken      { return Regex(s) }
func (PatternRuleBuilderHelper) CsRegex(s string) *PatternToken    { return CsRegex(s) }
