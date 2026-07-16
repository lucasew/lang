package ar

// Twin of ArabicTransVerbRuleTest — surface stem+clitic heuristic.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicTransVerbRule_Rule(t *testing.T) {
	rule := NewArabicTransVerbRule(nil)
	require.NotNil(t, rule)
	// Bare lemma forms should not flag
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("أفاض"))))
	// Attached-looking form without preposition (surface): أفاضه
	require.NotEqual(t, 0, len(rule.Match(languagetool.AnalyzePlain("أفاضه الماء"))))
}
