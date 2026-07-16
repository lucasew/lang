package ekavian

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleGrammarEkavianReplaceRule(t *testing.T) {
	rule := NewSimpleGrammarEkavianReplaceRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Плаћа у еуро."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "евро", matches[0].GetSuggestedReplacements()[0])
}
