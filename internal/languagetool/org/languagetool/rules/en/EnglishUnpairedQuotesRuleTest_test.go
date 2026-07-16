package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishUnpairedQuotesRuleTest.java
// Subset of cases that work without POS-tagged apostrophe exceptions.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishUnpairedQuotesRule_Rule(t *testing.T) {
	rule := NewEnglishUnpairedQuotesRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("This is a word 'test'."))
	require.Equal(t, 0, matchN("This is what he said: \"We believe in freedom. This is what we do.\""))
	// clear unpaired double quotes
	require.Equal(t, 1, matchN("\"I'm over here, she said."))
}
