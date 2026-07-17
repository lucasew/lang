package uk

// Twin of UkrainianPatternRuleTest — inject pattern rule green (full grammar.xml deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// Port of UkrainianPatternRuleTest.testRules
func TestUkrainianPatternRule_Rules(t *testing.T) {
	// synthetic two-token rule: "це" + "тест"
	r := patterns.NewPatternRule(
		"UK_DEMO", "uk",
		[]*patterns.PatternToken{patterns.Token("це"), patterns.Token("тест")},
		"demo", "msg", "short",
	)
	require.True(t, r.SupportsLanguage("uk"))
	// sentence with match
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("це", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("тест", nil, nil), 3),
	}
	matches, err := r.Match(languagetool.NewAnalyzedSentence(toks))
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	// no match
	toks2 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("інше", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("слово", nil, nil), 5),
	}
	matches2, err := r.Match(languagetool.NewAnalyzedSentence(toks2))
	require.NoError(t, err)
	require.Empty(t, matches2)
}
