package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikAsturianSpellerRule(t *testing.T) {
	r := NewMorfologikAsturianSpellerRule()
	// Java MorfologikAsturianSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_AST", MorfologikAsturianSpellerRuleID)
	require.Equal(t, "/ast/hunspell/ast_ES.dict", AsturianSpellerDict)
	require.Equal(t, MorfologikAsturianSpellerRuleID, r.GetID())
	require.Equal(t, AsturianSpellerDict, r.GetFileName())
}
