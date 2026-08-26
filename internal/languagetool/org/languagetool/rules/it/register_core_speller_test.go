package it

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Faithful: RegisterCore registers Java createDefaultSpellingRule / speller getId.
func TestRegisterCore_MorfologikSpellerID(t *testing.T) {
	lt := languagetool.NewJLanguageTool("it")
	RegisterCoreItalianRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_IT_IT")
}
