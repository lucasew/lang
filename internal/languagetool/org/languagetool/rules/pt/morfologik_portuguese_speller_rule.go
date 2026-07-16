package pt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

// Portuguese speller variants.
const (
	MorfologikPortuguesePTSpellerRuleID = "MORFOLOGIK_RULE_PT_PT"
	MorfologikPortugueseBRSpellerRuleID = "MORFOLOGIK_RULE_PT_BR"
	PortuguesePTDict                    = "/pt/hunspell/pt_PT.dict"
	PortugueseBRDict                    = "/pt/hunspell/pt_BR.dict"
)

// MorfologikPortugueseSpellerRule ports rules.pt.MorfologikPortugueseSpellerRule.
type MorfologikPortugueseSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	VariantCode string
}

func NewMorfologikPortugueseSpellerRule(variantCode, dict, id string) *MorfologikPortugueseSpellerRule {
	return &MorfologikPortugueseSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(id, "pt", dict, nil),
		VariantCode:           variantCode,
	}
}

func NewMorfologikPortugalPortugueseSpellerRule() *MorfologikPortugueseSpellerRule {
	return NewMorfologikPortugueseSpellerRule("pt-PT", PortuguesePTDict, MorfologikPortuguesePTSpellerRuleID)
}

func NewMorfologikBrazilianPortugueseSpellerRule() *MorfologikPortugueseSpellerRule {
	return NewMorfologikPortugueseSpellerRule("pt-BR", PortugueseBRDict, MorfologikPortugueseBRSpellerRuleID)
}
