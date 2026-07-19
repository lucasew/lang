package gl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreRules_NoSoftInventSequences(t *testing.T) {
	// Soft invent token sequences removed; official grammar.xml is incomplete until loaded.
	lt := languagetool.NewJLanguageTool("gl")
	RegisterCoreGalicianRules(lt)
	ids := lt.GetAllRegisteredRuleIDs()
	require.NotContains(t, ids, "GL_A_O")
}
