package uk

// RuleExceptionType ports RuleException.Type.
type RuleExceptionType int

const (
	RuleExceptionNone RuleExceptionType = iota
	RuleExceptionException
	RuleExceptionSkip
)

// RuleException ports org.languagetool.rules.uk.RuleException.
type RuleException struct {
	Type RuleExceptionType
	Skip int
}

func NewRuleException(t RuleExceptionType) RuleException {
	return RuleException{Type: t, Skip: 0}
}

func NewRuleExceptionSkip(skip int) RuleException {
	return RuleException{Type: RuleExceptionSkip, Skip: skip}
}
