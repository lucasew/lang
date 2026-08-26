package patterns

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDemoPatternRule_Rules(t *testing.T) {
	// Demo language pattern: foo bar
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("foo", nil, nil), 0),
		languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken("bar", nil, nil), 4),
	}
	sent := testSentence(toks...)
	rule := NewPatternRule("DEMO_RULE", "xx",
		[]*PatternToken{Token("foo"), Token("bar")},
		"demo", "found", "")
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

func TestDemoPatternRule_GrammarRulesFromXML2(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="DEMO_XML" name="from xml">
      <pattern><token>hello</token></pattern>
      <message>hi</message>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "demo.xml", "xx")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "DEMO_XML", rules[0].ID)
}

// makeDemoPatternRule ports PatternRuleTest.makePatternRule for Demo language.
func makeDemoPatternRule(s string, caseSensitive, regex bool) *PatternRule {
	parts := strings.Split(s, " ")
	var tokens []*PatternToken
	for _, element := range parts {
		pos := element == languagetool.SentenceStartTagName
		var pToken *PatternToken
		if !pos {
			pToken = NewPatternToken(element, caseSensitive, regex, false)
		} else {
			pToken = NewPatternToken("", caseSensitive, regex, false)
			pToken.SetPosToken(PosToken{PosTag: element, Regexp: false, Negate: false})
		}
		tokens = append(tokens, pToken)
	}
	return NewPatternRule("ID1", "xx", tokens, "test rule", "user visible message", "short comment")
}

// Twin of DemoPatternRuleTest.testMakeSuggestionUppercase
func TestDemoPatternRule_MakeSuggestionUppercase(t *testing.T) {
	// PatternToken "Were" case-insensitive → matches "Were"
	// Message suggestions "where" / "we" → capitalized at sentence start
	pt := NewPatternToken("Were", false, false, false)
	msg := "Did you mean: <suggestion>where</suggestion> or <suggestion>we</suggestion>?"
	rule := NewPatternRule("MY_ID", "xx", []*PatternToken{pt}, "desc", msg, "msg")
	sent := languagetool.AnalyzePlain("Were are in the process of ...")
	ms, err := rule.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	reps := ms[0].GetSuggestedReplacements()
	require.Len(t, reps, 2)
	require.Equal(t, "Where", reps[0])
	require.Equal(t, "We", reps[1])
}

// Twin of DemoPatternRuleTest.testRule
func TestDemoPatternRule_Rule(t *testing.T) {
	pr := makeDemoPatternRule("one", false, false)
	ms, err := pr.Match(languagetool.AnalyzePlain("A non-matching sentence."))
	require.NoError(t, err)
	require.Empty(t, ms)

	ms, err = pr.Match(languagetool.AnalyzePlain("A matching sentence with one match."))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, 25, ms[0].GetFromPos())
	require.Equal(t, 28, ms[0].GetToPos())
	require.Equal(t, -1, ms[0].GetColumn())
	require.Equal(t, -1, ms[0].GetLine())
	require.Equal(t, "ID1", ms[0].GetRule().(interface{ GetID() string }).GetID())
	require.Equal(t, "user visible message", ms[0].GetMessage())
	require.Equal(t, "short comment", ms[0].GetShortMessage())

	ms, err = pr.Match(languagetool.AnalyzePlain("one one and one: three matches"))
	require.NoError(t, err)
	require.Len(t, ms, 3)

	pr = makeDemoPatternRule("one two", false, false)
	ms, err = pr.Match(languagetool.AnalyzePlain("this is one not two"))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = pr.Match(languagetool.AnalyzePlain("this is two one"))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = pr.Match(languagetool.AnalyzePlain("this is one two three"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	ms, err = pr.Match(languagetool.AnalyzePlain("one two"))
	require.NoError(t, err)
	require.Len(t, ms, 1)

	pr = makeDemoPatternRule("one|foo|xxxx two", false, true)
	ms, err = pr.Match(languagetool.AnalyzePlain("one foo three"))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = pr.Match(languagetool.AnalyzePlain("one two"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	ms, err = pr.Match(languagetool.AnalyzePlain("foo two"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	ms, err = pr.Match(languagetool.AnalyzePlain("one foo two"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	ms, err = pr.Match(languagetool.AnalyzePlain("y x z one two blah foo"))
	require.NoError(t, err)
	require.Len(t, ms, 1)

	pr = makeDemoPatternRule("one|foo|xxxx two|yyy", false, true)
	ms, err = pr.Match(languagetool.AnalyzePlain("one, yyy"))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = pr.Match(languagetool.AnalyzePlain("one yyy"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	ms, err = pr.Match(languagetool.AnalyzePlain("xxxx two"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	ms, err = pr.Match(languagetool.AnalyzePlain("xxxx yyy"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

// Twin of DemoPatternRuleTest.testSentenceStart
func TestDemoPatternRule_SentenceStart(t *testing.T) {
	pr := makeDemoPatternRule("SENT_START One", false, false)
	ms, err := pr.Match(languagetool.AnalyzePlain("Not One word."))
	require.NoError(t, err)
	require.Empty(t, ms)
	ms, err = pr.Match(languagetool.AnalyzePlain("One word."))
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

// Twin of DemoPatternRuleTest.testFormatMultipleSynthesis
func TestDemoPatternRule_FormatMultipleSynthesis(t *testing.T) {
	suggestions1 := []string{"blah blah", "foo bar"}
	require.Equal(t,
		"This is how you should write: <suggestion>blah blah</suggestion>, <suggestion>foo bar</suggestion>.",
		formatMultipleSynthesis(suggestions1, "This is how you should write: <suggestion>", "</suggestion>."))
	suggestions2 := []string{"test", " "}
	require.Equal(t,
		"This is how you should write: <suggestion>test</suggestion>, <suggestion> </suggestion>.",
		formatMultipleSynthesis(suggestions2, "This is how you should write: <suggestion>", "</suggestion>."))
}
