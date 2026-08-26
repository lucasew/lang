package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestRegisterGrammarXML_VariantDefaultEnabledAfterDefaultOff(t *testing.T) {
	// Valencian enables EXIGEIX_VERBS_VALENCIANS; XML marks it default off.
	xml := `<?xml version="1.0"?>
<rules lang="ca">
  <category id="MISC" name="misc">
    <rule id="EXIGEIX_VERBS_VALENCIANS" name="val" default="off">
      <pattern>
        <token>foo</token>
      </pattern>
      <message>msg</message>
    </rule>
    <rule id="OTHER_OFF" name="other" default="off">
      <pattern>
        <token>bar</token>
      </pattern>
      <message>msg2</message>
    </rule>
  </category>
</rules>`

	lt := languagetool.NewJLanguageTool("ca-ES-valencia")
	n, err := patterns.RegisterGrammarXML(lt, xml, "test.xml", "ca")
	require.NoError(t, err)
	require.Equal(t, 2, n)

	// Variant enabled list re-enables EXIGEIX after MarkDefaultOff from XML.
	require.False(t, lt.IsRuleDisabled("EXIGEIX_VERBS_VALENCIANS"),
		"variant default enabled should setDefaultOn")
	require.Contains(t, lt.EnabledRules, "EXIGEIX_VERBS_VALENCIANS")
	// Unrelated default-off stays off.
	require.True(t, lt.IsRuleDisabled("OTHER_OFF"))
}

func TestRegisterGrammarXML_FrenchBE_DisablesDoubler(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="fr">
  <category id="MISC" name="misc">
    <rule id="DOUBLER_UNE_CLASSE" name="doubler">
      <pattern>
        <token>foo</token>
      </pattern>
      <message>msg</message>
    </rule>
  </category>
</rules>`
	lt := languagetool.NewJLanguageTool("fr-BE")
	_, err := patterns.RegisterGrammarXML(lt, xml, "test.xml", "fr")
	require.NoError(t, err)
	require.True(t, lt.IsRuleDisabled("DOUBLER_UNE_CLASSE"))
}
