package sk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikSlovakSpellerRuleID = "MORFOLOGIK_RULE_SK_SK"
	SlovakSpellerDict = "/sk/hunspell/sk_SK.dict"
)

type MorfologikSlovakSpellerRule struct { *morfologik.MorfologikSpellerRule }

func NewMorfologikSlovakSpellerRule() *MorfologikSlovakSpellerRule {
	r := &MorfologikSlovakSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(MorfologikSlovakSpellerRuleID, "sk", SlovakSpellerDict, nil),
	}
	// Java MorfologikSpellerRule.initSpeller when binary present.
	r.InitSpellersFromGetters(nil, nil)
	return r
}
