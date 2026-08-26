package ro

// Outcome twins for Romanian XmlRuleDisambiguator as used by Romanian.createDefaultDisambiguator:
// Java new XmlRuleDisambiguator(this) — useGlobalDisambiguation=false
// (language XML only; does NOT append disambiguation-global.xml).

import (
	"testing"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func TestRomanianXmlRuleDisambiguator_Load(t *testing.T) {
	require.NotEmpty(t, DiscoverRomanianDisambiguationXML(),
		"official ro/disambiguation.xml must be discoverable")
	xml := RomanianXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.IsType(t, &disambigrules.XmlRuleDisambiguator{}, xml)
	require.NotEmpty(t, xml.Rules)
	require.Same(t, xml, RomanianXmlRuleDisambiguator(), "process-cached singleton")
}
