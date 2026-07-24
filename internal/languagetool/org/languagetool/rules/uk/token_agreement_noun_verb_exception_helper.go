package uk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// TokenAgreementNounVerbExceptionHelper anchors the Java twin name.
// Logic lives in IsNounVerbException (token_agreement_core.go).
type TokenAgreementNounVerbExceptionHelper struct{}

func NewTokenAgreementNounVerbExceptionHelper() *TokenAgreementNounVerbExceptionHelper {
	return &TokenAgreementNounVerbExceptionHelper{}
}

// Exception ports isException(tokens, nounPos, verbPos).
func (TokenAgreementNounVerbExceptionHelper) Exception(tokens []*languagetool.AnalyzedTokenReadings, nounPos, verbPos int) bool {
	return IsNounVerbException(tokens, nounPos, verbPos)
}

// IsNonPluralA ports isNonPluralA (shared with adj-noun plural conj logic).
func (TokenAgreementNounVerbExceptionHelper) IsNonPluralA(tokens []*languagetool.AnalyzedTokenReadings, pos int) bool {
	return IsNonPluralA(tokens, pos)
}
