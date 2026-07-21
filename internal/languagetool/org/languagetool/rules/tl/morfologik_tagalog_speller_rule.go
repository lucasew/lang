package tl

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	// MorfologikTagalogSpellerRuleID ports MorfologikTagalogSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_TL"
	MorfologikTagalogSpellerRuleID = "MORFOLOGIK_RULE_TL"
	// TagalogSpellerDict ports MorfologikTagalogSpellerRule.getFileName().
	// Java: "/tl/hunspell/tl_PH.dict"
	TagalogSpellerDict = "/tl/hunspell/tl_PH.dict"
)

// MorfologikTagalogSpellerRule ports language.tl.MorfologikTagalogSpellerRule.
type MorfologikTagalogSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikTagalogSpellerRule() *MorfologikTagalogSpellerRule {
	r := &MorfologikTagalogSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikTagalogSpellerRuleID, "tl", TagalogSpellerDict, nil),
	}
	// Java MorfologikSpellerRule.initSpeller when binary present.
	r.InitSpellersFromGetters(nil, nil)
	return r
}
