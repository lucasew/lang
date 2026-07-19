package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestException_NegatePos(t *testing.T) {
	// exception postag="CD" negate_pos="yes" → exception matches when POS is NOT CD
	pt := NewPatternToken("the", false, false, false)
	pt.SetStringPosExceptionFullNeg("", false, false, false, "CD", false, true)
	m := NewPatternTokenMatcher(pt)

	cd := "CD"
	nn := "NN"
	tokCD := languagetool.NewAnalyzedToken("two", &cd, nil)
	tokNN := languagetool.NewAnalyzedToken("cat", &nn, nil)
	require.False(t, m.isExceptionMatchedCompletely(tokCD), "CD should not trigger negate_pos exception")
	require.True(t, m.isExceptionMatchedCompletely(tokNN), "NN should trigger negate_pos exception")
}

func TestException_SurfaceNegate(t *testing.T) {
	// exception surface "foo" negate=yes → exception matches when surface is NOT foo
	pt := NewPatternToken("x", false, false, false)
	pt.SetStringPosExceptionFullNeg("foo", false, false, true, "", false, false)
	m := NewPatternTokenMatcher(pt)
	tokFoo := languagetool.NewAnalyzedToken("foo", nil, nil)
	tokBar := languagetool.NewAnalyzedToken("bar", nil, nil)
	require.False(t, m.isExceptionMatchedCompletely(tokFoo))
	require.True(t, m.isExceptionMatchedCompletely(tokBar))
}

func TestException_LoaderNegatePos(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <category id="C" name="C">
    <rule id="R" name="R">
      <pattern>
        <token>the
          <exception postag="CD" negate_pos="yes"/>
        </token>
      </pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	pt := rules[0].PatternTokens[0]
	require.True(t, pt.TokenExceptionPosNegation)
	require.Equal(t, "CD", pt.TokenExceptionPosTag)
}
