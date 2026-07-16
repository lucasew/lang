package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatternRuleLoader_GetRules(t *testing.T) {
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
    <rule id="R2" name="Rule two">
      <pattern><token>x</token></pattern>
      <message>m2</message>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "test.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	require.Equal(t, "R1", rules[0].ID)
	require.Equal(t, "R2", rules[1].ID)
	require.Len(t, rules[0].PatternTokens, 2)
	require.True(t, rules[0].PatternTokens[1].Regexp)
}
