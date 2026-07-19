package de

// Twin of MissingCommaRelativeClauseRuleTest — Java uses POS/morph (no surface invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestMissingCommaRelativeClauseRule_Match(t *testing.T) {
	rule := NewMissingCommaRelativeClauseRule(nil)
	// untagged AnalyzePlain must not invent relative-comma hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das Auto das am Straßenrand steht parkt im Halteverbot."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Computer machen die Leute dumm."))))

	behind := NewMissingCommaRelativeClauseRuleBehind(nil)
	require.Equal(t, 0, len(behind.Match(languagetool.AnalyzePlain("Das Auto, das am Straßenrand steht parkt im Halteverbot."))))
	require.Equal(t, 0, len(behind.Match(languagetool.AnalyzePlain("Das Auto, das am Straßenrand steht, parkt im Halteverbot."))))

	// IDs match Java
	require.Equal(t, "COMMA_IN_FRONT_RELATIVE_CLAUSE", rule.GetID())
	require.Equal(t, "COMMA_BEHIND_RELATIVE_CLAUSE", behind.GetID())
}
