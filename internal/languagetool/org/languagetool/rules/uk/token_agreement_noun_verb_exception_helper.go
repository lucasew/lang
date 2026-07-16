package uk

// TokenAgreementNounVerbExceptionHelper ports org.languagetool.rules.uk.TokenAgreementNounVerbExceptionHelper (exception surface).
// Full dictionary-driven exception tables are deferred; callers can inject IsException.
type TokenAgreementNounVerbExceptionHelper struct {
	// IsException optional override for tests / full port.
	IsException func(tokens []string, a, b int) bool
}

func NewTokenAgreementNounVerbExceptionHelper() *TokenAgreementNounVerbExceptionHelper { return &TokenAgreementNounVerbExceptionHelper{} }

// Exception reports whether the pair at positions should be ignored.
func (h *TokenAgreementNounVerbExceptionHelper) Exception(tokens []string, a, b int) bool {
	if h != nil && h.IsException != nil {
		return h.IsException(tokens, a, b)
	}
	return false
}
