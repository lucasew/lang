package spelling

// RuleWithLanguage ports org.languagetool.rules.spelling.RuleWithLanguage.
// Language is a short code string until the Language type is fully ported.
type RuleWithLanguage struct {
	Rule         any
	LanguageCode string
}

func NewRuleWithLanguage(rule any, languageCode string) RuleWithLanguage {
	if rule == nil {
		panic("rule required")
	}
	if languageCode == "" {
		panic("language required")
	}
	return RuleWithLanguage{Rule: rule, LanguageCode: languageCode}
}

func (r RuleWithLanguage) GetRule() any            { return r.Rule }
func (r RuleWithLanguage) GetLanguageCode() string { return r.LanguageCode }

func (r RuleWithLanguage) Equal(o RuleWithLanguage) bool {
	return r.Rule == o.Rule && r.LanguageCode == o.LanguageCode
}
