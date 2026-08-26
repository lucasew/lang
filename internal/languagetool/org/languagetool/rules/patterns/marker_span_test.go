package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPatternRule_MarkerSpanOnly(t *testing.T) {
	// Java: <token>foo</token><marker><token>bar</token></marker> → match [bar only]
	xml := `<?xml version="1.0"?><rules lang="en"><category id="C" name="c">
		<rule id="M"><pattern>
			<token>foo</token>
			<marker><token>bar</token></marker>
		</pattern><message>m</message></rule>
	</category></rules>`
	loader := NewPatternRuleLoader()
	loader.SetRelaxedMode(true)
	ars, err := loader.GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.NotEmpty(t, ars)
	ar := ars[0]
	require.Len(t, ar.PatternTokens, 2)
	require.False(t, ar.PatternTokens[0].InsideMarker, "token outside marker")
	require.True(t, ar.PatternTokens[1].InsideMarker, "token inside marker")
	// Java prepareRule: startPos=1, endCorr=(1+1)-2 = 0
	require.Equal(t, 1, ar.StartPositionCorrection)
	require.Equal(t, 0, ar.EndPositionCorrection)

	pr := NewPatternRule(ar.ID, "en", ar.PatternTokens, "", "m", "")
	pr.StartPositionCorrection = ar.StartPositionCorrection
	pr.EndPositionCorrection = ar.EndPositionCorrection
	sent := languagetool.AnalyzePlain("foo bar")
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, 4, ms[0].FromPos)
	require.Equal(t, 7, ms[0].ToPos)
}

func TestPatternRuleLoader_SuggestionsOutMsgAndMarkerTail(t *testing.T) {
	// Outer <suggestion> → suggestionsOutMsg; marker mid-pattern → endCorr negative.
	xml := `<?xml version="1.0"?><rules lang="en"><category id="C" name="c">
		<rule id="S"><pattern>
			<marker><token>foo</token></marker>
			<token>bar</token>
		</pattern>
		<message>use this</message>
		<suggestion>baz</suggestion>
		</rule>
	</category></rules>`
	loader := NewPatternRuleLoader()
	loader.SetRelaxedMode(true)
	ars, err := loader.GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.NotEmpty(t, ars)
	ar := ars[0]
	require.Contains(t, ar.SuggestionsOutMsg, "baz")
	require.Equal(t, 0, ar.StartPositionCorrection) // marker at start
	require.Equal(t, -1, ar.EndPositionCorrection)  // one token after marker

	pr := NewPatternRule(ar.ID, "en", ar.PatternTokens, ar.Description, ar.Message, ar.ShortMessage)
	pr.SuggestionsOutMsg = ar.SuggestionsOutMsg
	pr.StartPositionCorrection = ar.StartPositionCorrection
	pr.EndPositionCorrection = ar.EndPositionCorrection
	sent := languagetool.AnalyzePlain("foo bar")
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	// Marker only on foo
	require.Equal(t, 0, ms[0].FromPos)
	require.Equal(t, 3, ms[0].ToPos)
	require.Contains(t, ms[0].GetSuggestedReplacements(), "baz")
}

func TestPatternRule_NoMarkerFullSpan(t *testing.T) {
	xml := `<?xml version="1.0"?><rules lang="en"><category id="C" name="c">
		<rule id="F"><pattern>
			<token>foo</token>
			<token>bar</token>
		</pattern><message>m</message></rule>
	</category></rules>`
	loader := NewPatternRuleLoader()
	loader.SetRelaxedMode(true)
	ars, err := loader.GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	pr := NewPatternRule(ars[0].ID, "en", ars[0].PatternTokens, "", "m", "")
	// Without <marker>, NewPatternToken defaults InsideMarker true → full span.
	for _, pt := range ars[0].PatternTokens {
		require.True(t, pt.InsideMarker)
	}
	sent := languagetool.AnalyzePlain("foo bar")
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].FromPos)
	require.Equal(t, 7, ms[0].ToPos)
}
