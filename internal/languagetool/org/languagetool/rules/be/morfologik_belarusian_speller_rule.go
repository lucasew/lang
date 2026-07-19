package be

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikBelarusianSpellerRuleID   = "MORFOLOGIK_RULE_BE_BY"
	MorfologikBelarusianSpellerRuleDict = "/be/hunspell/be_BY.dict"
)

// MorfologikBelarusianSpellerRule ports rules.be.MorfologikBelarusianSpellerRule.
type MorfologikBelarusianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikBelarusianSpellerRule() *MorfologikBelarusianSpellerRule {
	r := &MorfologikBelarusianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikBelarusianSpellerRuleID, "be", MorfologikBelarusianSpellerRuleDict, nil),
	}
	// Java isLatinScript() = false
	if r.SpellingCheckRule != nil {
		r.NonLatinScript = true
	}
	return r
}

// IsLatinScript ports isLatinScript.
func (r *MorfologikBelarusianSpellerRule) IsLatinScript() bool { return false }
