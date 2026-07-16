package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFalseFriendRuleLoader(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rulegroup id="ABILITY">
    <rule>
      <pattern lang="en">
        <token inflected="yes">ability</token>
      </pattern>
      <translation lang="fr">aptitude</translation>
    </rule>
    <rule>
      <pattern lang="fr">
        <token inflected="yes">habileté</token>
      </pattern>
      <translation lang="en">skill</translation>
    </rule>
  </rulegroup>
</rules>`
	loader := NewFalseFriendRuleLoader("FF: {0} means {1} ({2})", "")
	rules, err := loader.GetRulesFromString(xml, "en", "fr")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "ABILITY", rules[0].ID)
	require.Contains(t, rules[0].Message, "ability")
	require.Contains(t, rules[0].Message, "aptitude")
	require.Contains(t, loader.SuggestionMap["ABILITY"], "aptitude")
}
