package en

// Twin of English WordRepeatRuleTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestWordRepeatRule_Rule(t *testing.T) {
	r := rules.NewWordRepeatRule(map[string]string{"repetition": "Repetition"})
	require.Empty(t, r.Match(languagetool.AnalyzePlain("This is fine")))
	require.Len(t, r.Match(languagetool.AnalyzePlain("This this is bad")), 1)
	// Known name repetition ignored by core rule
	require.Empty(t, r.Match(languagetool.AnalyzePlain("Duran Duran")))
}
