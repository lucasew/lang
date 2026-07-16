package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of ConfusionSetLoaderTest.testConfusionSetLoading (standalone resource path)
func TestConfusionSetLoader_languagetool_standalone_ConfusionSetLoading(t *testing.T) {
	raw := `
# comment
their; there; 10
you're; your; 5
car -> cars; 2
`
	m, err := NewConfusionSetLoader(nil).LoadConfusionPairs(strings.NewReader(raw))
	require.NoError(t, err)
	require.Contains(t, m, "their")
	require.Contains(t, m, "there")
	require.Contains(t, m, "you're")
	require.Contains(t, m, "car")
}
