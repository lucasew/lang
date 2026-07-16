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
