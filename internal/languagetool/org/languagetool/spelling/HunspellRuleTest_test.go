package spelling

// Twin of languagetool-standalone HunspellRuleTest — MapHunspell inject with alt-lang surface.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/stretchr/testify/require"
)

// Port of HunspellRuleTest.testRuleWithGermanAndAltLang
func TestHunspellRule_RuleWithGermanAndAltLang(t *testing.T) {
	// primary DE dict
	de := hunspell.NewMapHunspellDictionary([]string{"Haus", "Hund", "und", "der"})
	// alt-lang EN words accepted via ignore / secondary lookup soft
	enWords := map[string]struct{}{"the": {}, "dog": {}, "house": {}}
	r := hunspell.NewHunspellRule("de-DE", de)
	// accept EN tokens via IgnoreWords (alt language soft path)
	for w := range enWords {
		r.AddIgnoreWords(w)
	}

	// pure German OK
	m, err := r.Match(languagetool.AnalyzePlain("Haus und Hund"))
	require.NoError(t, err)
	require.Empty(t, m)

	// German misspelling still flagged
	m, err = r.Match(languagetool.AnalyzePlain("Huas"))
	require.NoError(t, err)
	require.NotEmpty(t, m)

	// EN word ignored (alt lang soft)
	m, err = r.Match(languagetool.AnalyzePlain("the dog"))
	require.NoError(t, err)
	require.Empty(t, m)
}
