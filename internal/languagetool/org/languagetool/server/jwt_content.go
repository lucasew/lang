package server

// JwtContent ports org.languagetool.server.JwtContent.
type JwtContent struct {
	IsValid  bool
	IsPremium bool
	Claims   map[string]any
}

// JwtNone is the empty/invalid JWT content sentinel.
var JwtNone = JwtContent{IsValid: false, IsPremium: false, Claims: map[string]any{}}

func NewJwtContent(isValid, isPremium bool, claims map[string]any) JwtContent {
	if claims == nil {
		claims = map[string]any{}
	}
	return JwtContent{IsValid: isValid, IsPremium: isPremium, Claims: claims}
}
