package sv

// Example-based twin for Swedish WordCoherencyRule.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWordCoherencyRule(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// Java example: mejl vs mail
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Jag skickar mejl varje dag."),
	})))
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Det är en blandning av mejl och mail i det du skriver."),
	})))
}
