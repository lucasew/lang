package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPatternToken_PhraseName(t *testing.T) {
	pt := NewPatternToken("hello", false, false, false)
	require.False(t, pt.IsPartOfPhrase())
	pt.SetPhraseName("GREET")
	require.True(t, pt.IsPartOfPhrase())
	require.Equal(t, "GREET", pt.GetPhraseName())
}

func TestPatternRule_ElementNoFromPhrase(t *testing.T) {
	// Two-token phrase + one normal token → elementNo [2, 1], useList true.
	a := NewPatternToken("good", false, false, false)
	a.SetPhraseName("G")
	b := NewPatternToken("morning", false, false, false)
	b.SetPhraseName("G")
	c := NewPatternToken("sir", false, false, false)
	pr := NewPatternRule("R", "en", []*PatternToken{a, b, c}, "d", "m", "")
	require.True(t, pr.UseList)
	require.Equal(t, []int{2, 1}, pr.GetElementNo())
}

func TestPatternRule_ElementNoNoPhrase(t *testing.T) {
	a := NewPatternToken("a", false, false, false)
	b := NewPatternToken("b", false, false, false)
	pr := NewPatternRule("R", "en", []*PatternToken{a, b}, "d", "m", "")
	require.False(t, pr.UseList)
	// Java still records 1 per non-phrase token
	require.Equal(t, []int{1, 1}, pr.GetElementNo())
}

func TestPatternRuleLoader_PhraserefSetsPhraseName(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <phrases>
    <phrase id="PAIR">
      <token>foo</token>
      <token>bar</token>
    </phrase>
  </phrases>
  <category>
    <rule id="P1" name="phrase">
      <pattern>
        <phraseref idref="PAIR"/>
        <token>baz</token>
      </pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 1)
	require.Len(t, ars[0].PatternTokens, 3)
	require.Equal(t, "PAIR", ars[0].PatternTokens[0].GetPhraseName())
	require.Equal(t, "PAIR", ars[0].PatternTokens[1].GetPhraseName())
	require.Equal(t, "", ars[0].PatternTokens[2].GetPhraseName())
	require.Equal(t, "baz", ars[0].PatternTokens[2].Token)

	pr := NewPatternRule(ars[0].ID, "en", ars[0].PatternTokens, ars[0].Description, ars[0].Message, "")
	require.True(t, pr.UseList)
	require.Equal(t, []int{2, 1}, pr.ElementNo)

	// Match "foo bar baz"
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("foo", 0), atr("bar", 4), atr("baz", 8),
	}
	ms, err := pr.Match(testSentence(toks...))
	require.NoError(t, err)
	require.Len(t, ms, 1)
}

func TestTranslateElementNo_PhraseSkip(t *testing.T) {
	// skip=1 with useList and elementNo[0]=2 → translate to 2 tokens.
	a := NewPatternToken("a", false, false, false)
	a.SetPhraseName("P")
	b := NewPatternToken("b", false, false, false)
	b.SetPhraseName("P")
	c := NewPatternToken("c", false, false, false)
	c.SetSkipNext(1) // skip one XML element (the phrase after? actually skip after c)
	// Pattern: optional setup - just test translate on matcher
	pr := NewPatternRule("R", "en", []*PatternToken{a, b, c}, "d", "m", "")
	m := NewPatternRuleMatcherFromPattern(pr)
	require.True(t, m.UseList)
	require.Equal(t, 2, m.translateElementNo(1)) // sum elementNo[0]
	require.Equal(t, 3, m.translateElementNo(2)) // 2+1
	require.Equal(t, -1, m.translateElementNo(-1))
	require.Equal(t, 0, m.translateElementNo(0))
}

func TestFormatMatches_PhraseLen(t *testing.T) {
	// Phrase of two tokens: elementNo [2], \1 synthesizes both with space.
	toks := []*languagetool.AnalyzedTokenReadings{
		atr("good", 0),
		atr("morning", 5),
	}
	m := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
	m.SetInMessageOnly(true)
	// positions: one entry per pattern token matcher (Java)
	// For phraseLen on index 0 = 2, tokenIndex = firstMatchTok+repTokenPos
	// j=0, repTokenPos=positions[0]=1 → tokenIndex=1, idx=0; synthesizes tokens 0 and 1
	ctx := PhraseMatchContext{UseList: true, ElementNo: []int{2}}
	msg := FormatMatches(toks, []int{1, 1}, 0, `Say \1`, []*Match{m}, "en", ctx)
	require.Equal(t, "Say good morning", msg)
}

func TestProcessElement_AdjustsPhraseMatchRef(t *testing.T) {
	// Second token in phrase with TokenMatch ref 1 → adjusted to 1+1-1=1? counter=1 → tokRef+0
	// counter=1 (second token): newRef = tokRef + 1 - 1 = tokRef. No change when tokRef=1?
	// Java: tokRef + counter - 1; counter starts 0.
	// First token counter=0: not adjusted (counter > 0 false)
	// Second counter=1: newRef = tokRef + 0 = tokRef if... wait tokRef+1-1=tokRef
	// Third counter=2: newRef = tokRef + 1
	a := NewPatternToken("x", false, false, false)
	a.SetPhraseName("P")
	b := NewPatternToken(`\1`, false, false, false)
	b.SetPhraseName("P")
	mm := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
	mm.SetTokenRef(1)
	b.SetMatch(mm)
	c := NewPatternToken(`\1`, false, false, false)
	c.SetPhraseName("P")
	mm2 := NewMatch("", "", false, "", "", CaseNone, false, false, IncludeNone)
	mm2.SetTokenRef(1)
	c.SetMatch(mm2)

	processElement([]*PatternToken{a, b, c})
	require.Equal(t, 1, b.GetMatch().GetTokenRef()) // 1+1-1=1
	require.Equal(t, 2, c.GetMatch().GetTokenRef()) // 1+2-1=2
	require.Equal(t, `\2`, c.Token)
}
