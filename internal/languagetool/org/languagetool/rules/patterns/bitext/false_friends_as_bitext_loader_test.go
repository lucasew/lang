package bitext

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFalseFriendsAsBitextLoader_HintsForPolishTranslators(t *testing.T) {
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
	loader := NewFalseFriendsAsBitextLoader()
	rules, err := loader.GetFalseFriendsAsBitext(
		strings.NewReader(xml),
		strings.NewReader(xml),
		"en", "pl",
	)
	require.NoError(t, err)
	require.NotEmpty(t, rules)
	require.Equal(t, "ACTUAL", rules[0].GetID())
	require.Equal(t, "en", rules[0].GetSourceLanguage())
}
