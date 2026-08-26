package uk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// TokenAgreementPrepNounExceptionHelper anchors the Java twin name.
// Logic lives in GetPrepNounExceptionStrong / NonInfl / Infl and IsPrepNounException
// (token_agreement_core.go), matching TokenAgreementPrepNounExceptionHelper.
type TokenAgreementPrepNounExceptionHelper struct{}

func NewTokenAgreementPrepNounExceptionHelper() *TokenAgreementPrepNounExceptionHelper {
	return &TokenAgreementPrepNounExceptionHelper{}
}

// GetExceptionStrong ports getExceptionStrong.
func (TokenAgreementPrepNounExceptionHelper) GetExceptionStrong(tokens []*languagetool.AnalyzedTokenReadings, i int, prepTok *languagetool.AnalyzedTokenReadings) RuleException {
	return GetPrepNounExceptionStrong(tokens, i, prepTok)
}

// GetExceptionNonInfl ports getExceptionNonInfl.
func (TokenAgreementPrepNounExceptionHelper) GetExceptionNonInfl(tokens []*languagetool.AnalyzedTokenReadings, i int) RuleException {
	return GetPrepNounExceptionNonInfl(tokens, i)
}

// GetExceptionInfl ports getExceptionInfl.
func (TokenAgreementPrepNounExceptionHelper) GetExceptionInfl(tokens []*languagetool.AnalyzedTokenReadings, prepPos, i int) RuleException {
	return GetPrepNounExceptionInfl(tokens, prepPos, i)
}

// Exception is a boolean any-hit wrapper (tests / soft callers).
func (TokenAgreementPrepNounExceptionHelper) Exception(tokens []*languagetool.AnalyzedTokenReadings, prepPos, nounPos int) bool {
	return IsPrepNounException(tokens, prepPos, nounPos)
}
