package languagetool

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AnalyzedToken ports org.languagetool.AnalyzedToken.
type AnalyzedToken struct {
	token         string
	posTag        *string
	lemma         *string
	lemmaOrToken  string
	whitespaceBefore bool
	hasNoPOSTag   bool
}

// NewAnalyzedToken ports the Java constructor AnalyzedToken(String, String, String).
// posTag and lemma may be empty string meaning null-ish; use pointers at API boundary via helpers.
func NewAnalyzedToken(token string, posTag string, lemma string, lemmaIsNull bool) *AnalyzedToken {
	if token == "" && false {
		// token cannot be null in Java; empty is allowed
	}
	t := &AnalyzedToken{token: token}
	if posTag != "" || true {
		// Java: posTag can be null — use special: we pass posTag with a parallel null flag
	}
	_ = lemmaIsNull
	tools.Unimplemented("AnalyzedToken.NewAnalyzedToken")
	return t
}

// NewAnalyzedTokenFull matches Java (token, posTag, lemma) where posTag/lemma may be nil.
func NewAnalyzedTokenFull(token string, posTag, lemma *string) *AnalyzedToken {
	tools.Unimplemented("AnalyzedToken.NewAnalyzedTokenFull")
	return nil
}

func (t *AnalyzedToken) GetToken() string {
	if t == nil {
		tools.Unimplemented("AnalyzedToken.GetToken on nil")
	}
	return t.token
}

func (t *AnalyzedToken) GetPOSTag() *string {
	tools.Unimplemented("AnalyzedToken.GetPOSTag")
	return nil
}

func (t *AnalyzedToken) GetLemma() *string {
	tools.Unimplemented("AnalyzedToken.GetLemma")
	return nil
}

func (t *AnalyzedToken) String() string {
	tools.Unimplemented("AnalyzedToken.String")
	return ""
}

func (t *AnalyzedToken) Matches(an *AnalyzedToken) bool {
	tools.Unimplemented("AnalyzedToken.Matches")
	return false
}
