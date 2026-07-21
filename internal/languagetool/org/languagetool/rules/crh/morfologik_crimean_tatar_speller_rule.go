package crh

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikCrimeanTatarSpellerRuleID   = "MORFOLOGIK_RULE_CRH_UA"
	MorfologikCrimeanTatarSpellerRuleDict = "/crh/hunspell/crh_UA.dict"
)

// MorfologikCrimeanTatarSpellerRule ports rules.crh.MorfologikCrimeanTatarSpellerRule.
type MorfologikCrimeanTatarSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikCrimeanTatarSpellerRule() *MorfologikCrimeanTatarSpellerRule {
	r := &MorfologikCrimeanTatarSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikCrimeanTatarSpellerRuleID, "crh", MorfologikCrimeanTatarSpellerRuleDict, nil),
	}
	// Java isLatinScript() = false
	if r.SpellingCheckRule != nil {
		r.NonLatinScript = true
	}
	// Java MorfologikSpellerRule.initSpeller when binary present.
	r.InitSpellersFromGetters(nil, nil)
	return r
}

// IsLatinScript ports isLatinScript.
func (r *MorfologikCrimeanTatarSpellerRule) IsLatinScript() bool { return false }
