package ast

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"

const (
	// MorfologikAsturianSpellerRuleID ports MorfologikAsturianSpellerRule.getId().
	// Java: "MORFOLOGIK_RULE_AST"
	MorfologikAsturianSpellerRuleID = "MORFOLOGIK_RULE_AST"
	// AsturianSpellerDict ports MorfologikAsturianSpellerRule.getFileName().
	// Java: "/ast/hunspell/ast_ES.dict"
	AsturianSpellerDict = "/ast/hunspell/ast_ES.dict"
)

// MorfologikAsturianSpellerRule ports language.rules.ast.MorfologikAsturianSpellerRule.
type MorfologikAsturianSpellerRule struct {
	*morfologik.MorfologikSpellerRule
}

func NewMorfologikAsturianSpellerRule() *MorfologikAsturianSpellerRule {
	return &MorfologikAsturianSpellerRule{
		MorfologikSpellerRule: morfologik.NewMorfologikSpellerRule(
			MorfologikAsturianSpellerRuleID, "ast", AsturianSpellerDict, nil),
	}
}
