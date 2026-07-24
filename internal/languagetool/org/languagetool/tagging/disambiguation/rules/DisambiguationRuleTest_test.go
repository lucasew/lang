package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisambiguationRule_DisambiguationRulesFromXML(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="CD" name="Tag numbers">
    <pattern>
      <token regexp="yes">\d+</token>
    </pattern>
    <disambig postag="CD"/>
  </rule>
</rules>`
	loaded, err := NewDisambiguationRuleLoader().GetRulesFromString(xml, "en", "test.xml")
	require.NoError(t, err)
	require.NotEmpty(t, loaded)
	require.Equal(t, "CD", loaded[0].ID)
}
