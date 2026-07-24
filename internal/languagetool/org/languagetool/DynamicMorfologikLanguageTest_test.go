package languagetool

// Twin of DynamicMorfologikLanguageTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDynamicMorfologikLanguage_Test(t *testing.T) {
	d := NewDynamicMorfologikLanguage("Custom", "xx", "/tmp/custom.dict")
	require.Equal(t, "Custom", d.Name)
	require.Equal(t, "xx", d.Code)
	require.Equal(t, "/tmp/custom.dict", d.SpellerDictPath())
	require.Equal(t, "XX_SPELLER_RULE", d.SpellerRuleID())
	require.Equal(t, []string{"XX_SPELLER_RULE"}, d.RelevantSpellerRuleIDs())
}
