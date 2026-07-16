package de

// Twin of MissingCommaRelativeClauseRuleTest (surface relative-pronoun heuristics).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMissingCommaRelativeClauseRule_Match(t *testing.T) {
	rule := NewMissingCommaRelativeClauseRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Das Auto das am Straßenrand steht parkt im Halteverbot."))
	require.Equal(t, 1, matchN("Die Frau die vor dem Auto steht hat schwarze Haare."))
	require.Equal(t, 1, matchN("Alles was ich habe, ist ein Buch."))
	require.Equal(t, 0, matchN("Computer machen die Leute dumm."))
	require.Equal(t, 0, matchN("Die Studenten, deren Urteil am stärksten von dem der Profis abwich, waren sich sicher."))

	behind := NewMissingCommaRelativeClauseRuleBehind(nil)
	require.Equal(t, 1, len(behind.Match(languagetool.AnalyzePlain("Das Auto, das am Straßenrand steht parkt im Halteverbot."))))
	require.Equal(t, 0, len(behind.Match(languagetool.AnalyzePlain("Das Auto, das am Straßenrand steht, parkt im Halteverbot."))))
}
