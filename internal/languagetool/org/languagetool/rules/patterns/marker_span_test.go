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

	pr := NewPatternRule(ar.ID, "en", ar.PatternTokens, "", "m", "")
	sent := languagetool.AnalyzePlain("foo bar")
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, 4, ms[0].FromPos)
	require.Equal(t, 7, ms[0].ToPos)
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
