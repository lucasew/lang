package ekavian

// Twin of languagetool-language-modules/sr/src/test/java/.../SimpleStyleEkavianReplaceRuleTest.java
// Java twin is empty; exercise dictionary surface matches.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleStyleEkavianReplaceRule_GetMessage(t *testing.T) {
	rule := NewSimpleStyleEkavianReplaceRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Купио је компјутер."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "рачунар", matches[0].GetSuggestedReplacements()[0])
}
