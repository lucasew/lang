package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPatternToken_SetPosMatch_CompileAndMatch(t *testing.T) {
	// Second token: <match no="0" setpos postag_replace from first token gender/case>
	// Simplified: first has POS "N:f:sg:acc", second requires POS derived to match ".*:f:sg:acc.*"
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="SETPOS" name="setpos demo">
      <pattern>
        <token postag="N:f:sg:acc"/>
        <token postag_regexp="yes">
          <match no="0" postag="N:([fm]):(sg|pl):(acc|nom)" postag_regexp="yes" postag_replace="N:$1:$2:$3" setpos="yes"/>
        </token>
      </pattern>
      <message>ok</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 1)
	require.True(t, ars[0].PatternTokens[1].IsReferenceElement())
	require.True(t, ars[0].PatternTokens[1].TokenMatch.SetsPos())
	require.Equal(t, 0, ars[0].PatternTokens[1].TokenMatch.GetTokenRef())

	// Build sentence: two tokens with matching POS via setpos
	pos1, lem1 := "N:f:sg:acc", "a"
	pos2, lem2 := "N:f:sg:acc", "b"
	// second token POS after setpos should be N:f:sg:acc from first
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("la", &pos1, &lem1), 0),
		languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("maison", &pos2, &lem2), 3),
	}
	pr := NewPatternRule(ars[0].ID, "en", ars[0].PatternTokens, ars[0].Description, ars[0].Message, "")
	ms, err := pr.Match(testSentence(toks...))
	require.NoError(t, err)
	require.Len(t, ms, 1, "setpos should align second POS with first")

	// Mismatch gender on second
	pos2m := "N:m:sg:acc"
	toks2 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("la", &pos1, &lem1), 0),
		languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken("mur", &pos2m, &lem2), 3),
	}
	ms2, err := pr.Match(languagetool.NewAnalyzedSentence(toks2))
	require.NoError(t, err)
	require.Empty(t, ms2, "setpos POS must not invent agreement")
}

func TestGetTargetPosTag_Replace(t *testing.T) {
	m := NewMatch("N:([fm]):(sg):(acc)", "N:$1:$2:$3", true, "", "", CaseNone, true, false, IncludeNone)
	pos, lem := "N:f:sg:acc", "x"
	atr := languagetool.NewAnalyzedTokenReadingsAt(
		languagetool.NewAnalyzedToken("x", &pos, &lem), 0)
	ms := NewMatchState(m)
	ms.SetToken(atr)
	got := ms.GetTargetPosTag()
	require.Equal(t, "N:f:sg:acc", got)
}
