package km

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"

// KhmerHunspellClasspath ports Java KhmerHunspellRule / Hunspell path for km_KH.
const KhmerHunspellClasspath = "/km/hunspell/km_KH.dic"

// KhmerHunspellRule ports org.languagetool.rules.km.KhmerHunspellRule.
type KhmerHunspellRule struct {
	*hunspell.HunspellRule
}

func NewKhmerHunspellRule(dict hunspell.HunspellDictionary) *KhmerHunspellRule {
	// Java KhmerHunspellRule extends HunspellRule without getId override → HUNSPELL_RULE.
	r := hunspell.NewHunspellRule("km", dict)
	// Java isLatinScript() = false
	if r != nil && r.SpellingCheckRule != nil {
		r.NonLatinScript = true
	}
	return &KhmerHunspellRule{HunspellRule: r}
}

// NewKhmerHunspellRuleDefault opens official km_KH.dic when present; nil fails closed.
func NewKhmerHunspellRuleDefault() *KhmerHunspellRule {
	return NewKhmerHunspellRule(hunspell.TryOpenFromClasspath(KhmerHunspellClasspath))
}

// IsLatinScript ports KhmerHunspellRule.isLatinScript() → false.
func (r *KhmerHunspellRule) IsLatinScript() bool { return false }
