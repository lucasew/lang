package ast

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	MorfologikAsturianSpellerRuleID   = "MORFOLOGIK_RULE_AST"
	MorfologikAsturianSpellerRuleDict = "/ast/hunspell/ast_ES.dict"
)

// MorfologikAsturianSpellerRule ports language.rules.ast.MorfologikAsturianSpellerRule.
type MorfologikAsturianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikAsturianSpellerRule() *MorfologikAsturianSpellerRule {
	return &MorfologikAsturianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikAsturianSpellerRuleID, "ast", MorfologikAsturianSpellerRuleDict, nil),
	}
}
