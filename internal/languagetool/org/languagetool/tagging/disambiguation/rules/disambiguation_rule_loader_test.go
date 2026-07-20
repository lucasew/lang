package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisambiguationRuleLoader(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="CD" name="Tag numbers">
    <pattern>
      <token regexp="yes">\d+</token>
    </pattern>
    <disambig postag="CD"/>
  </rule>
</rules>`
	rules, err := NewDisambiguationRuleLoader().GetRulesFromString(xml, "xx", "disambiguation.xml")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "CD", rules[0].ID)
	require.Equal(t, "CD", rules[0].DisambiguatedPOS)
	require.Equal(t, ActionReplace, rules[0].Action)
	require.True(t, rules[0].Tokens[0].Regexp)
}

func TestDisambiguationRuleLoader_PostagToken(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="DT_NN" name="det then noun">
    <pattern>
      <token>the</token>
      <token postag="NN"/>
    </pattern>
    <disambig action="filter" postag="NN"/>
  </rule>
</rules>`
	rules, err := NewDisambiguationRuleLoader().GetRulesFromString(xml, "en", "t.xml")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "the", rules[0].Tokens[0].Token)
	require.NotNil(t, rules[0].Tokens[1].Pos)
	require.Equal(t, "NN", rules[0].Tokens[1].Pos.PosTag)
}

func TestDisambigLoader_MatchElement(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="M1" name="match filter">
    <pattern>
      <token>will</token>
      <marker><token>run</token></marker>
    </pattern>
    <disambig action="filter">
      <match no="2" postag="VB.*" postag_regexp="yes"/>
    </disambig>
  </rule>
  <rule id="M2" name="lemma filter">
    <pattern>
      <token>foo</token>
    </pattern>
    <disambig action="filter">
      <match no="1" postag="noun:inanim:m:v_rod">рік</match>
    </disambig>
  </rule>
</rules>`
	ars, err := NewDisambiguationRuleLoader().GetRulesFromString(xml, "en", "t.xml")
	require.NoError(t, err)
	require.Len(t, ars, 2)

	require.Equal(t, ActionFilter, ars[0].Action)
	require.NotNil(t, ars[0].MatchElement)
	require.Equal(t, 2, ars[0].MatchElement.GetTokenRef())
	require.True(t, ars[0].MatchElement.IsPostagRegexp())
	require.Equal(t, "VB.*", ars[0].MatchElement.GetPosTag())
	require.True(t, ars[0].MatchElement.HasPosRegexp())

	require.NotNil(t, ars[1].MatchElement)
	require.Equal(t, 1, ars[1].MatchElement.GetTokenRef())
	require.True(t, ars[1].MatchElement.IsStaticLemma())
	require.Equal(t, "рік", ars[1].MatchElement.GetLemma())
	require.Equal(t, "noun:inanim:m:v_rod", ars[1].MatchElement.GetPosTag())
}
