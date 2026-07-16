package diff

// Twin of LightRuleMatchParserTest
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLightRuleMatchParser_Parse(t *testing.T) {
	s := "1.) Line 1, column 9, Rule ID: EN_A_VS_AN premium: false\n" +
		"Message: Use 'a' instead of 'an'\n" +
		"Suggestion: a\n" +
		"This is an test. \n" +
		"        ^^       \n" +
		"Tags: [picky, fake]\n" +
		"Time: 10ms for 1 sentences (0.7 sentences/sec)\n"
	matches := NewLightRuleMatchParser().ParseOutput(strings.NewReader(s)).Matches
	require.Len(t, matches, 1)
	m := matches[0]
	require.Equal(t, 1, m.Line)
	require.Equal(t, 9, m.Column)
	require.Equal(t, "EN_A_VS_AN", m.GetFullRuleID())
	require.Equal(t, "EN_A_VS_AN", m.GetRuleID())
	require.False(t, m.Premium)
	require.Equal(t, []string{"a"}, m.Suggestions)
	require.Equal(t, "Use 'a' instead of 'an'", m.Message)
	require.Equal(t, "an", m.CoveredText)
	require.Empty(t, m.RuleSource)
	require.Equal(t, "This is <span class='marker'>an</span> test. ", m.Context)
	require.Empty(t, m.Title)
}

func TestLightRuleMatchParser_ParseTwoMatches(t *testing.T) {
	s := "1.) Line 1, column 9, Rule ID: EN_A_VS_AN premium: false\n" +
		"Message: Use 'a' instead of 'an'\n" +
		"Suggestion: a\n" +
		"This is an test. \n" +
		"        ^^       \n" +
		"\n" +
		"2.) Line 5, column 6, Rule ID: FOO2 premium: true\n" +
		"Message: message2\n" +
		"Suggestion: something\n" +
		"This is somethink test. \n" +
		"        ^^^^^^^^^       \n" +
		"Time: 10ms for 1 sentences (0.7 sentences/sec)\n"
	matches := NewLightRuleMatchParser().ParseOutput(strings.NewReader(s)).Matches
	require.Len(t, matches, 2)
	require.Equal(t, "EN_A_VS_AN", matches[0].GetRuleID())
	require.False(t, matches[0].Premium)
	require.Equal(t, []string{"a"}, matches[0].Suggestions)
	require.Equal(t, "an", matches[0].CoveredText)
	require.Equal(t, 5, matches[1].Line)
	require.Equal(t, 6, matches[1].Column)
	require.Equal(t, "FOO2", matches[1].GetRuleID())
	require.True(t, matches[1].Premium)
	require.Equal(t, []string{"something"}, matches[1].Suggestions)
	require.Equal(t, "somethink", matches[1].CoveredText)
	require.Equal(t, "This is <span class='marker'>somethink</span> test. ", matches[1].Context)
}

func TestLightRuleMatchParser_ParseNightlyFormat(t *testing.T) {
	s := "Title: Anarchism\n" +
		"Line 1, column 35, Rule ID: EN_QUOTES[1] premium: false\n" +
		"Message: Use a smart opening quote here: '“'.\n" +
		"Suggestion: “\n" +
		"Rule source: /org/languagetool/rules/en/grammar.xml\n" +
		"Proponents of anarchism, known as \"anarchists\", advocate stateless societies based on...\n" +
		"                                  ^                                                  \n"
	matches := NewLightRuleMatchParser().ParseOutput(strings.NewReader(s)).Matches
	require.Len(t, matches, 1)
	m := matches[0]
	require.Equal(t, 1, m.Line)
	require.Equal(t, 35, m.Column)
	require.Equal(t, "EN_QUOTES[1]", m.GetFullRuleID())
	require.Equal(t, "EN_QUOTES", m.GetRuleID())
	require.Equal(t, "1", m.GetSubID())
	require.Equal(t, []string{"“"}, m.Suggestions)
	require.Equal(t, "\"", m.CoveredText)
	require.Equal(t, "/org/languagetool/rules/en/grammar.xml", m.RuleSource)
	require.Equal(t, "Anarchism", m.Title)
	require.Contains(t, m.Context, "<span class='marker'>\"</span>")
}

func TestLightRuleMatchParser_ParseNightlyFormatNoSuggestion(t *testing.T) {
	s := "Title: Anarchism\n" +
		"Line 1, column 35, Rule ID: EN_QUOTES[1] premium: false\n" +
		"Message: Use a smart opening quote here: '“'.\n" +
		"Rule source: /org/languagetool/rules/en/grammar-testme.xml\n" +
		"Proponents of anarchism, known as \"anarchists\", advocate stateless societies based on...\n" +
		"                                  ^                                                  \n"
	matches := NewLightRuleMatchParser().ParseOutput(strings.NewReader(s)).Matches
	require.Len(t, matches, 1)
	m := matches[0]
	require.Equal(t, "EN_QUOTES", m.GetRuleID())
	require.Equal(t, []string{"null"}, m.Suggestions)
	require.Equal(t, "/org/languagetool/rules/en/grammar-testme.xml", m.RuleSource)
	require.Equal(t, "Anarchism", m.Title)
	require.Equal(t, "\"", m.CoveredText)
}
