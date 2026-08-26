package patterns

import (
	"regexp"
	"strings"
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

func TestRegexPatternRule_RequiredSubstringsAndUTF16(t *testing.T) {
	re := regexp.MustCompile(`hello world`)
	rule := NewRegexPatternRule("R", "d", "m", "", "", "en", re, 0)
	require.NotNil(t, rule.RequiredSubstrings)
	require.Contains(t, rule.RequiredSubstrings.String(), "hello")

	// Over UTF-16 MaxSentLength → no matches (Java text.length()).
	// Use BMP runes so rune count == UTF-16 length.
	long := strings.Repeat("a", MaxSentLength+1)
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(long, nil, nil))
	sent := testSentence(tok)
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Empty(t, matches)
}

func TestRegexPatternRule_ProcessMessageMatchReplace(t *testing.T) {
	re := regexp.MustCompile(`(foo)`)
	rule := NewRegexPatternRule("R", "d", `use <suggestion>\1</suggestion>`, "", "", "en", re, 0)
	// Java processMessage: Match with regexReplace on backref
	mw := NewMatch("", "", false, "f(.*)", "F$1", CaseAllUpper, false, false, IncludeNone)
	rule.SuggestionMatches = []*Match{mw}
	tok := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("foo", nil, nil))
	sent := testSentence(tok)
	matches, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	// replace f(.*) → F$1 on "foo" → "Foo", then ALLUPPER → "FOO"
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "FOO")
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
