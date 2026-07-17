package gl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreGalicianRules_Patterns(t *testing.T) {
	lt := languagetool.NewJLanguageTool("gl")
	RegisterCoreGalicianRules(lt)
	m := lt.Check("Vou a o mercado.")
	found := false
	for _, x := range m {
		if x.RuleID == "GL_A_O" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
