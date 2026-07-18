package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPatternRuleLoader(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="R1" name="Rule one">
      <pattern>
        <token>foo</token>
        <token regexp="yes">b.r</token>
      </pattern>
      <message>bad</message>
      <short>s</short>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "test.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "R1", rules[0].ID)
	require.Equal(t, "Rule one", rules[0].Description)
	require.Len(t, rules[0].PatternTokens, 2)
	require.Equal(t, "foo", rules[0].PatternTokens[0].Token)
	require.True(t, rules[0].PatternTokens[1].Regexp)
	require.Equal(t, "bad", rules[0].Message)

	// round-trip match
	pr := NewPatternRule(rules[0].ID, "en", rules[0].PatternTokens, rules[0].Description, rules[0].Message, rules[0].ShortMessage)
	require.NotNil(t, pr)
}

func TestPatternRuleLoaderRelaxed(t *testing.T) {
	xml := `<rules><category><rule><pattern><token>x</token></pattern></rule></category></rules>`
	l := NewPatternRuleLoader()
	l.SetRelaxedMode(true)
	rules, err := l.GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 1)
}

func TestPatternRuleLoader_ExceptionAndInflected(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="EX1" name="with exception">
      <pattern>
        <token inflected="yes">run<exception>running</exception></token>
        <token>fast</token>
      </pattern>
      <message>x</message>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	pt := rules[0].PatternTokens[0]
	require.True(t, pt.MatchInflected)
	require.Equal(t, "running", pt.TokenException)
	require.False(t, pt.TokenExceptionRE)

	m := NewPatternTokenMatcher(pt)
	runTok := languagetool.NewAnalyzedToken("run", nil, strPtr("run"))
	runningTok := languagetool.NewAnalyzedToken("running", nil, strPtr("run"))
	// Java isMatched is surface/POS only; exceptions apply via
	// isExceptionMatchedCompletely after any reading matches (IsMatchedReadings).
	require.True(t, m.IsMatched(runTok))
	require.True(t, m.IsMatched(runningTok), "lemma run still matches pattern before exception gate")
	require.True(t, m.IsMatchedReadings(languagetool.NewAnalyzedTokenReadings(runTok)))
	require.False(t, m.IsMatchedReadings(languagetool.NewAnalyzedTokenReadings(runningTok)),
		"surface exception running blocks via isExceptionMatchedCompletely")
}

func TestPatternRuleLoader_PreviousNextException(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category id="C" name="C">
    <rule id="PREV" name="prev">
      <pattern>
        <token>mine<exception scope="previous">not</exception></token>
      </pattern>
      <message>m</message>
      <short>s</short>
    </rule>
    <rule id="NEXT" name="next">
      <pattern>
        <token>can<exception scope="next" regexp="yes">be|do</exception></token>
      </pattern>
      <message>m</message>
      <short>s</short>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	require.Equal(t, "not", rules[0].PatternTokens[0].PreviousException)
	require.Equal(t, "be|do", rules[1].PatternTokens[0].NextException)
	require.True(t, rules[1].PatternTokens[0].NextExceptionRE)
}
