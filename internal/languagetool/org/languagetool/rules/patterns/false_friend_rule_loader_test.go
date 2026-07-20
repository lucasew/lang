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
	// Java formatTranslations wraps translations in quotes
	require.Contains(t, rules[0].Message, `"aptitude"`)
	require.Contains(t, loader.SuggestionMap["ABILITY"], "aptitude")
	require.True(t, rules[0].Tokens[0].MatchInflected)
}

func TestFalseFriendRuleLoader_PostagNegate(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rulegroup id="ACCORD">
    <rule>
      <pattern lang="en">
        <token inflected="yes" postag="NN.*" postag_regexp="yes">accord</token>
      </pattern>
      <translation lang="fr">accord</translation>
    </rule>
    <rule>
      <pattern lang="fr">
        <token negate="yes">na</token>
        <token>accord</token>
      </pattern>
      <translation lang="en">chord</translation>
    </rule>
  </rulegroup>
</rules>`
	loader := NewFalseFriendRuleLoader("", "")
	// EN text, FR mother: first rule only
	rules, err := loader.GetRulesFromString(xml, "en", "fr")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.NotNil(t, rules[0].Tokens[0].Pos)
	require.Equal(t, "NN.*", rules[0].Tokens[0].Pos.PosTag)
	require.True(t, rules[0].Tokens[0].Pos.Regexp)

	// FR text, EN mother: second rule with negate token
	rulesFR, err := loader.GetRulesFromString(xml, "fr", "en")
	require.NoError(t, err)
	require.Len(t, rulesFR, 1)
	require.True(t, rulesFR[0].Tokens[0].Negation)
	require.Equal(t, "na", rulesFR[0].Tokens[0].Token)
}

// Empty constructor args use MessagesBundle_en false_friend_* defaults (not invent 2-arg).
func TestFalseFriendRuleLoader_MessagesBundleDefaults(t *testing.T) {
	loader := NewFalseFriendRuleLoader("", "")
	require.Equal(t, messagesFalseFriendHint, loader.FalseFriendHint)
	require.Equal(t, messagesFalseFriendSugg, loader.FalseFriendSugg)
	xml := `<?xml version="1.0"?>
<rules>
  <rulegroup id="ABILITY">
    <rule>
      <pattern lang="en"><token>ability</token></pattern>
      <translation lang="fr">aptitude</translation>
    </rule>
  </rulegroup>
</rules>`
	rules, err := loader.GetRulesFromString(xml, "en", "fr")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	// MessagesBundle_en: Hint: "{0}" ({1}) means {2} ({3}).
	require.Contains(t, rules[0].Message, "Hint:")
	require.Contains(t, rules[0].Message, "English")
	require.Contains(t, rules[0].Message, "French")
	require.Contains(t, rules[0].Message, `"aptitude"`)
	// Distinct surface → Did you mean {0}?
	require.Contains(t, rules[0].Message, "Did you mean")
	require.NotContains(t, rules[0].Message, "Possible false friend")
}

// false-friends.xml has skip="-1" on PL pracować; load into PatternToken.SkipNext.
func TestFalseFriendRuleLoader_Skip(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rulegroup id="WORK">
    <rule>
      <pattern lang="pl">
        <token inflected="yes" skip="-1">pracować</token>
      </pattern>
      <translation lang="en">work</translation>
    </rule>
  </rulegroup>
</rules>`
	loader := NewFalseFriendRuleLoader("", "")
	rules, err := loader.GetRulesFromString(xml, "pl", "en")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, -1, rules[0].Tokens[0].SkipNext)
}
