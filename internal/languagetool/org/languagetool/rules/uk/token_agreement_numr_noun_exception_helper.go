package uk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// TokenAgreementNumrNounExceptionHelper anchors the Java twin name.
// Logic lives in IsNumrNounException (token_agreement_core.go).
type TokenAgreementNumrNounExceptionHelper struct{}

func NewTokenAgreementNumrNounExceptionHelper() *TokenAgreementNumrNounExceptionHelper {
	return &TokenAgreementNumrNounExceptionHelper{}
}

// Exception ports isException(tokens, numrPos, nounPos) boolean form.
func (TokenAgreementNumrNounExceptionHelper) Exception(tokens []*languagetool.AnalyzedTokenReadings, numrPos, nounPos int) bool {
	return IsNumrNounException(tokens, numrPos, nounPos)
}
