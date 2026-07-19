package el

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikGreekSpellerRuleID = "MORFOLOGIK_RULE_EL_GR"
	GreekSpellerDict             = "/el/hunspell/el_GR.dict"
)

// MorfologikGreekSpellerRule ports rules.el.MorfologikGreekSpellerRule.
type MorfologikGreekSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikGreekSpellerRule() *MorfologikGreekSpellerRule {
	r := &MorfologikGreekSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikGreekSpellerRuleID, "el", GreekSpellerDict, nil),
	}
	// Java isLatinScript() = false
	if r.SpellingCheckRule != nil {
		r.NonLatinScript = true
	}
	return r
}

// IsLatinScript ports isLatinScript.
func (r *MorfologikGreekSpellerRule) IsLatinScript() bool { return false }
