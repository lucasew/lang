package server

// ActiveRulesMBean ports org.languagetool.server.ActiveRulesMBean.
type ActiveRulesMBean interface {
	GetActivePatternRules() map[string]int
	GetActiveSpellChecks() []string
}

// Ensure ActiveRules implements ActiveRulesMBean.
var _ ActiveRulesMBean = (*ActiveRules)(nil)
