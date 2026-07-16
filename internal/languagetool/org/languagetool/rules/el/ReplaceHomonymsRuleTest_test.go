package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/ReplaceHomonymsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestReplaceHomonymsRule_Rule(t *testing.T) {
	rule := NewReplaceHomonymsRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Στην Ελλάδα επικρατεί εύκρατο κλίμα."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Καλή τύχη σου εύχομαι."))))
}

func TestReplaceHomonymsRule_RuleInsideOfSentence(t *testing.T) {
	rule := NewReplaceHomonymsRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Του ευχήθηκα καλή τείχη για το διαγώνισμα."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "καλή τύχη", matches[0].GetSuggestedReplacements()[0])
}
