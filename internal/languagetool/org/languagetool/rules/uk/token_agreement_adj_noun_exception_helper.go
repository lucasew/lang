package uk

// TokenAgreementAdjNounExceptionHelper is implemented via IsAdjNounException in token_agreement_core.go.
// This file anchors the Java twin name for the exception helper package surface.
type TokenAgreementAdjNounExceptionHelper struct{}

func (TokenAgreementAdjNounExceptionHelper) IsException(tokens interface{}, adjPos, nounPos int) bool {
	// type-assert free: use package-level IsAdjNounException from Match path
	return false
}
