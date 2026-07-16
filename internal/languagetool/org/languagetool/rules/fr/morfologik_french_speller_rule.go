package fr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

const (
	MorfologikFrenchSpellerRuleID = "MORFOLOGIK_RULE_FR"
	FrenchSpellerDict             = "/fr/hunspell/fr.dict"
)

// MorfologikFrenchSpellerRule ports rules.fr.MorfologikFrenchSpellerRule.
type MorfologikFrenchSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikFrenchSpellerRule() *MorfologikFrenchSpellerRule {
	return &MorfologikFrenchSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikFrenchSpellerRuleID, "fr", FrenchSpellerDict, nil),
	}
}
