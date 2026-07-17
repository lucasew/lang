package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRussianRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ru")
	RegisterCoreRussianRules(lt)
	m := lt.Check("Он пошёл в в дом.")
	found := false
	for _, x := range m {
		if x.RuleID == "RU_В_В" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
