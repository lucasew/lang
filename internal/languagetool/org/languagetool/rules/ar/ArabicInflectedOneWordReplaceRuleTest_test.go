package ar

// Twin of ArabicInflectedOneWordReplaceRuleTest (surface clitic/stem matching).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicInflectedOneWordReplaceRule_Rule(t *testing.T) {
	rule := NewArabicInflectedOneWordReplaceRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	// Correct
	require.Equal(t, 0, matchN("أجريت بحوثا في المخبر"))
	require.Equal(t, 0, matchN("وجعل لكم من أزواجكم بنين وحفدة"))
	// Errors (inflected / with proclitic)
	require.NotEqual(t, 0, matchN("أجريت أبحاثا في المخبر"))
	require.NotEqual(t, 0, matchN("وجعل لكم من أزواجكم بنين وأحفاد"))
}
