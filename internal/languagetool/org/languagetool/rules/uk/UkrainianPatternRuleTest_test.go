package uk

// Twin of UkrainianPatternRuleTest — inject pattern rule green (full grammar.xml deferred).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

// ukTestSentence prepends SENT_START like Java AnalyzedSentence / PatternRuleMatcher tests.
// Without it AbstractTokenBasedRule.canBeIgnoredFor drops short token lists (Java minTokenCount).
func ukTestSentence(toks ...*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedSentence {
	ss := languagetool.SentenceStartTagName
	start := languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("", &ss, nil), 0)
	all := make([]*languagetool.AnalyzedTokenReadings, 0, len(toks)+1)
	all = append(all, start)
	all = append(all, toks...)
	return languagetool.NewAnalyzedSentence(all)
}

// Port of UkrainianPatternRuleTest.testRules (full runGrammarRulesFromXmlTest deferred).
// Synthetic two-token rule exercises matcher with Java-shaped sentence (SENT_START + tokens).
func TestUkrainianPatternRule_Rules(t *testing.T) {
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
	matches, err := r.Match(ukTestSentence(toks...))
	require.NoError(t, err)
	require.NotEmpty(t, matches)
	// no match
	toks2 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("інше", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("слово", nil, nil), 5),
	}
	matches2, err := r.Match(ukTestSentence(toks2...))
	require.NoError(t, err)
	require.Empty(t, matches2)
}
