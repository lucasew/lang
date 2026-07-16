package km

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"

// KhmerHunspellRule ports org.languagetool.rules.km.KhmerHunspellRule.
type KhmerHunspellRule struct {
	*hunspell.HunspellRule
}

func NewKhmerHunspellRule(dict hunspell.HunspellDictionary) *KhmerHunspellRule {
	r := hunspell.NewHunspellRule("km", dict)
	// Java uses HunspellRule base id; keep language-specific override optional.
	r.ID = "HUNSPELL_RULE_KM"
	return &KhmerHunspellRule{HunspellRule: r}
}

func NewKhmerHunspellRuleDefault() *KhmerHunspellRule {
	return NewKhmerHunspellRule(nil)
}
