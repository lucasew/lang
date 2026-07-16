package uk

// TokenAgreementNumrNounExceptionHelper ports org.languagetool.rules.uk.TokenAgreementNumrNounExceptionHelper (exception surface).
// Full dictionary-driven exception tables are deferred; callers can inject IsException.
type TokenAgreementNumrNounExceptionHelper struct {
	// IsException optional override for tests / full port.
	IsException func(tokens []string, a, b int) bool
}

func NewTokenAgreementNumrNounExceptionHelper() *TokenAgreementNumrNounExceptionHelper { return &TokenAgreementNumrNounExceptionHelper{} }

// Exception reports whether the pair at positions should be ignored.
func (h *TokenAgreementNumrNounExceptionHelper) Exception(tokens []string, a, b int) bool {
	if h != nil && h.IsException != nil {
		return h.IsException(tokens, a, b)
	}
	return false
}
