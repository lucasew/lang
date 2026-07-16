package patterns

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFalseFriendRule_HintsForEnglishSpeakers(t *testing.T) {
	// en text, fr mother tongue → en→fr false friends for French mother tongue speakers learning EN
	xml := `<?xml version="1.0"?>
<rules>
  <rulegroup id="GIFT">
    <rule>
      <pattern lang="en"><token>gift</token></pattern>
      <translation lang="fr">présent</translation>
    </rule>
    <rule>
      <pattern lang="fr"><token>gift</token></pattern>
      <translation lang="en">poison</translation>
    </rule>
  </rulegroup>
</rules>`
	loader := NewFalseFriendRuleLoader("FF: {0} means {1} ({2})", "")
	// mother tongue fr, text language en
	rules, err := loader.GetRulesFromString(xml, "en", "fr")
	require.NoError(t, err)
	require.NotEmpty(t, rules)
}

func TestFalseFriendRule_HintsForPolishSpeakers(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rulegroup id="ACTUAL">
    <rule>
      <pattern lang="en"><token>actual</token></pattern>
      <translation lang="pl">aktualny</translation>
    </rule>
    <rule>
      <pattern lang="pl"><token>aktualny</token></pattern>
      <translation lang="en">current</translation>
    </rule>
  </rulegroup>
</rules>`
	loader := NewFalseFriendRuleLoader("FF: {0} means {1} ({2})", "")
	rules, err := loader.GetRulesFromString(xml, "en", "pl")
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	require.Equal(t, "ACTUAL", rules[0].ID)
}
