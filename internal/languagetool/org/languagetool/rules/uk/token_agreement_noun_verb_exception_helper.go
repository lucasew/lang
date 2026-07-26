package uk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// TokenAgreementNounVerbExceptionHelper ports TokenAgreementNounVerbExceptionHelper.
// Full isException body lives in IsNounVerbException (token_agreement_core.go) to share
// helpers with other agreement rules; this type is the Java twin name anchor.
type TokenAgreementNounVerbExceptionHelper struct{}

func NewTokenAgreementNounVerbExceptionHelper() *TokenAgreementNounVerbExceptionHelper {
	return &TokenAgreementNounVerbExceptionHelper{}
}

// Exception ports isException(tokens, nounPos, verbPos, …).
// Java also passes master/slave inflections and selected readings; those are recomputed
// inside IsNounVerbException from the token positions (same readings Java would pass).
func (TokenAgreementNounVerbExceptionHelper) Exception(tokens []*languagetool.AnalyzedTokenReadings, nounPos, verbPos int) bool {
	return IsNounVerbException(tokens, nounPos, verbPos)
}

// IsNonPluralA ports TokenAgreementNounVerbExceptionHelper.isNonPluralA
// (shared with adj-noun plural conj logic).
func (TokenAgreementNounVerbExceptionHelper) IsNonPluralA(tokens []*languagetool.AnalyzedTokenReadings, pos int) bool {
	return IsNonPluralA(tokens, pos)
}

// HasMascFemLemma ports TokenAgreementNounVerbExceptionHelper.hasMascFemLemma.
func (TokenAgreementNounVerbExceptionHelper) HasMascFemLemma(tok *languagetool.AnalyzedTokenReadings) bool {
	return HasMascFemLemma(tok)
}
