package de

// Twin of SimilarNameRuleTest (surface name heuristic).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimilarNameRule_Rule(t *testing.T) {
	rule := NewSimilarNameRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 1, matchN("Hier steht Angela Müller. Im nächsten Satz dann Miller."))
	require.Equal(t, 0, matchN("Hier steht Angela Müller. Im nächsten Satz dann Müllers Ehemann."))
	require.Equal(t, 0, matchN("Hier steht Angela Müller. Dann Mulla, nicht ähnlich genug."))
	require.Equal(t, 0, matchN("Ein Mikrocontroller, bei Mikrocontrollern"))
	require.Equal(t, 0, matchN("Hier steht das Rad Deiner Freundin. Und Deinem Hund geht es gut?"))
}
