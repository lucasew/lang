package language

// Twin of SerbianTest.getRuleFileNames
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerbian_GetRuleFileNames(t *testing.T) {
	want := []string{
		"/org/languagetool/rules/sr/grammar.xml",
		"/org/languagetool/rules/sr/grammar-barbarism.xml",
		"/org/languagetool/rules/sr/grammar-logical.xml",
		"/org/languagetool/rules/sr/grammar-punctuation.xml",
		"/org/languagetool/rules/sr/grammar-spelling.xml",
		"/org/languagetool/rules/sr/grammar-style.xml",
	}
	require.Equal(t, want, NewSerbian().GetRuleFileNames())
	require.Equal(t, "sr", DefaultSerbian.GetShortCode())
	require.Equal(t, "Serbian", DefaultSerbian.GetName())
}
