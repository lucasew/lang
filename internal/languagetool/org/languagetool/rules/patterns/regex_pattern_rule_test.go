package patterns

import (
	"regexp"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegexPatternRuleMatch(t *testing.T) {
	// fo[ou] bar — mark whole match (group 0)
	re := regexp.MustCompile(`fo[ou] bar`)
	rule := NewRegexPatternRule(
		"REGEX_DEMO", "demo",
		`msg: <suggestion>a suggestion \0</suggestion>`,
		"short",
		`<suggestion>another suggestion bar</suggestion>`,
		"en", re, 0,
	)
	// build sentence text via tokens that concatenate to the string
	// AnalyzedSentence.GetText() joins tokens — check how
	text := "This is foo bar and fou bar"
	// use a single-token sentence if GetText uses token strings
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(text, nil, nil))
	sent := testSentence(tok)
	// if GetText doesn't equal full text, set via whatever API exists
	gotText := sent.GetText()
	if gotText != text {
		t.Logf("GetText()=%q; using direct tokens", gotText)
	}
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 2)
	require.Equal(t, 8, matches[0].FromPos)
	require.Equal(t, 15, matches[0].ToPos)
	require.Contains(t, matches[0].Message, "suggestion")
	require.NotEmpty(t, matches[0].GetSuggestedReplacements())
}

func TestRegexPatternRuleMarkGroup(t *testing.T) {
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

func TestPatternRuleTransformer(t *testing.T) {
	a := NewAbstractPatternRule("A", "a", "en", nil, false)
	b := NewAbstractPatternRule("B", "b", "en", nil, false)
	tr := PatternRuleTransformerFunc(func(rules []*AbstractPatternRule) TransformedRules {
		return NewTransformedRules(rules[1:], []any{"wrapped:" + rules[0].ID})
	})
	rem, out := TransformPatternRules([]*AbstractPatternRule{a, b}, tr)
	require.Len(t, rem, 1)
	require.Equal(t, "B", rem[0].ID)
	require.Equal(t, []any{"wrapped:A"}, out)
}
