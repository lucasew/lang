package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreArabicRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ar")
	RegisterCoreArabicRules(lt)
	m := lt.Check("ذهبت في في البيت.")
	found := false
	for _, x := range m {
		if x.RuleID == "AR_FI_FI" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
