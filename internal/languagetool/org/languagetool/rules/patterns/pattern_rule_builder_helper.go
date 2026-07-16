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
