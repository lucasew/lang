package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCore_RussianYODefaultOff(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	RegisterCoreRussianRules(lt)
	require.Contains(t, lt.GetAllRegisteredRuleIDs(), "MORFOLOGIK_RULE_RU_RU_YO")
	require.Contains(t, lt.GetDefaultOffRuleIDs(), "MORFOLOGIK_RULE_RU_RU_YO")
}
