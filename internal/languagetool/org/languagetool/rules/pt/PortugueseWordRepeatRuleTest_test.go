package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPortugueseWordRepeatRule_Ignore(t *testing.T) {
	rule := NewPortugueseWordRepeatRule(map[string]string{"repetition": "Repetição"})
	ignoreAt := func(input string, position int) bool {
		t.Helper()
		sent := languagetool.AnalyzePlain(input)
		tokens := sent.GetTokensWithoutWhitespace()
		return rule.Ignore(tokens, position)
	}
	require.False(t, ignoreAt("no repetition", 2))
	require.True(t, ignoreAt("blá blá", 2))
	require.True(t, ignoreAt("Aaptos aaptos", 2))
	require.True(t, ignoreAt("Logo logo vamos ao mercado", 2))
	// With WordTokenizer, coloquem-na splits; ignore at 2 is "-" not a na-na pair.
	// Java v0.13 keeps coloquem-na as one token so position 2 is second "na" without hyphen path.
	// Twin intent: do not ignore bare "na na" as hyphenated clitic without hyphen.
	require.False(t, ignoreAt("Coloquem-na na sala.", 2))
}
