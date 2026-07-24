package gl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestGLAdvancedSynthesizerFilterRegistered(t *testing.T) {
	class := "org.languagetool.rules.gl.AdvancedSynthesizerFilter"
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	require.NotNil(t, patterns.GlobalRuleFilterCreator.GetFilter(class))
}
