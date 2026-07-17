package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreUkrainianRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("uk")
	RegisterCoreUkrainianRules(lt)
	m := lt.Check("Він пішов в в дім.")
	found := false
	for _, x := range m {
		if x.RuleID == "UK_В_В" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
