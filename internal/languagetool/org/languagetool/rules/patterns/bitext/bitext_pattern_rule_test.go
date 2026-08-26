package bitext

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

type alwaysMatch struct {
	id string
	ms []*rules.RuleMatch
}

func (a alwaysMatch) GetID() string          { return a.id }
func (a alwaysMatch) GetDescription() string { return a.id }
func (a alwaysMatch) GetMessage() string     { return "msg" }
func (a alwaysMatch) Match(s *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	return a.ms, nil
}

func TestBitextPatternRule(t *testing.T) {
	srcSent := languagetool.AnalyzePlain("gift")
	trgSent := languagetool.AnalyzePlain("Gift")
	hit := rules.NewRuleMatch(rules.NewFakeRule("FF"), trgSent, 0, 4, "false friend")
	src := alwaysMatch{id: "FF", ms: []*rules.RuleMatch{hit}}
	trg := alwaysMatch{id: "FF", ms: []*rules.RuleMatch{hit}}
	r := NewBitextPatternRule(src, trg)
	require.Equal(t, "FF", r.GetID())
	got := r.MatchBitext(srcSent, trgSent)
	require.Len(t, got, 1)

	// no src match
	src2 := alwaysMatch{id: "FF", ms: nil}
	r2 := NewBitextPatternRule(src2, trg)
	require.Empty(t, r2.MatchBitext(srcSent, trgSent))
}

func TestBitextPatternRuleLoader(t *testing.T) {
	xml := `<?xml version="1.0"?>
	<bitextrules>
	  <rule id="B1" name="test">
	    <pattern lang="en"><token>foo</token></pattern>
	    <pattern lang="de"><token>bar</token></pattern>
	    <message>msg</message>
	  </rule>
	</bitextrules>`
	rules, err := NewBitextPatternRuleLoader().GetRules(strings.NewReader(xml), "test.xml")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "B1", rules[0].GetID())
	require.NotNil(t, rules[0].GetSrcRule())
	_, ok := rules[0].GetSrcRule().(*patterns.PatternRule)
	require.True(t, ok)
}
