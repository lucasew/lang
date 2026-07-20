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
	require.Len(t, pt.CurrentExceptions, 1)
}

// Java isExceptionMatched: multi current exceptions are a disjunction.
func TestException_MultiCurrent(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <category id="C" name="C">
    <rule id="R" name="R">
      <pattern>
        <token>word
          <exception>foo</exception>
          <exception>bar</exception>
          <exception postag="CD"/>
        </token>
      </pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	pt := rules[0].PatternTokens[0]
	require.Len(t, pt.CurrentExceptions, 3)
	require.Equal(t, "foo", pt.TokenException, "first exception mirrored to legacy fields")

	m := NewPatternTokenMatcher(pt)
	require.True(t, m.isExceptionMatchedCompletely(languagetool.NewAnalyzedToken("foo", nil, nil)))
	require.True(t, m.isExceptionMatchedCompletely(languagetool.NewAnalyzedToken("bar", nil, nil)))
	cd := "CD"
	require.True(t, m.isExceptionMatchedCompletely(languagetool.NewAnalyzedToken("two", &cd, nil)))
	require.False(t, m.isExceptionMatchedCompletely(languagetool.NewAnalyzedToken("baz", nil, nil)))

	// IsMatchedReadings: pattern surface "run" blocked when exception also matches surface.
	pt3 := NewPatternToken("run", false, false, false)
	pt3.AddCurrentException(NewPatternToken("run", false, false, false))
	pt3.AddCurrentException(NewPatternToken("ran", false, false, false))
	m3 := NewPatternTokenMatcher(pt3)
	runOK := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("run", nil, nil))
	require.False(t, m3.IsMatchedReadings(runOK), "current exception matching surface blocks")
	// different surface matches neither pattern nor exception gate after pattern fail
	other := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("walk", nil, nil))
	require.False(t, m3.IsMatchedReadings(other))
}

// Java isMatched: (text ^ neg) && (pos ^ posNeg) — both negations independent.
func TestIsMatched_NegationAndPosNegationXOR(t *testing.T) {
	pt := NewPatternToken("one", false, false, false)
	pt.SetNegation(true)
	pt.SetPosToken(PosToken{PosTag: "CD", Regexp: false, Negate: true})
	m := NewPatternTokenMatcher(pt)
	// "one"/CD: (true^true)=false && (true^true)=false → false
	cd := "CD"
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("one", &cd, nil)))
	// "two"/NN: (false^true)=true && (false^true)=true → true
	nn := "NN"
	require.True(t, m.IsMatched(languagetool.NewAnalyzedToken("two", &nn, nil)))
	// "two"/CD: (false^true)=true && (true^true)=false → false
	require.False(t, m.IsMatched(languagetool.NewAnalyzedToken("two", &cd, nil)))
}
