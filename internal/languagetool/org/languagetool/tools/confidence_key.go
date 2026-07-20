package tools

// ConfidenceKey ports org.languagetool.tools.ConfidenceKey.
// Language is a short-code string until Language is fully ported.
type ConfidenceKey struct {
	LanguageCode string
	RuleID       string
}

func NewConfidenceKey(languageCode, ruleID string) ConfidenceKey {
	// Java: Objects.requireNonNull(lang/ruleId) — empty string is allowed.
	return ConfidenceKey{LanguageCode: languageCode, RuleID: ruleID}
}

func (k ConfidenceKey) String() string {
	return k.LanguageCode + "/" + k.RuleID
}

func (k ConfidenceKey) Equal(o ConfidenceKey) bool {
	return k.LanguageCode == o.LanguageCode && k.RuleID == o.RuleID
}
