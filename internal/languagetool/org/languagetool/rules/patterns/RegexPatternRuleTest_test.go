package patterns

// Twin of languagetool-core/src/test/java/org/languagetool/rules/patterns/RegexPatternRuleTest.java
import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of RegexPatternRuleTest.testMatch
func TestRegexPatternRule_Match(t *testing.T) {
	re := regexp.MustCompile(`fo[ou] bar`)
	rule := NewRegexPatternRule(
		"REGEX_DEMO", "demo",
		`msg: <suggestion>a suggestion \0</suggestion>`,
		"short",
		`<suggestion>another suggestion bar</suggestion>`,
		"en", re, 0,
	)
	text := "This is foo bar and fou bar"
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(text, nil, nil))
	sent := testSentence(tok)
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 2)
}

// Port of RegexPatternRuleTest.testMatchWithMark
func TestRegexPatternRule_MatchWithMark(t *testing.T) {
	re := regexp.MustCompile(`(fo[ou]) bar`)
	rule := NewRegexPatternRule("M", "d", "m", "", "", "en", re, 1)
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("This is foo bar", nil, nil))
	sent := testSentence(tok)
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	require.Equal(t, 8, matches[0].FromPos)
	require.Equal(t, 11, matches[0].ToPos)
}
