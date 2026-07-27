package da

// Outcome twins for Danish XmlRuleDisambiguator as used by Danish.createDefaultDisambiguator:
// Java new XmlRuleDisambiguator(this) — useGlobalDisambiguation=false
// (language XML only; does NOT append disambiguation-global.xml).

import (
	"testing"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func TestDanishXmlRuleDisambiguator_Load(t *testing.T) {
	require.NotEmpty(t, DiscoverDanishDisambiguationXML(),
		"official da/disambiguation.xml must be discoverable")
	xml := DanishXmlRuleDisambiguator()
	require.NotNil(t, xml)
	require.IsType(t, &disambigrules.XmlRuleDisambiguator{}, xml)
	require.NotEmpty(t, xml.Rules)
	// Official DA pack: sub-pron (2), sub-ver (3), ver-sub (7), pron-sub (1) → 13 rules.
	require.GreaterOrEqual(t, len(xml.Rules), 13)
	require.Same(t, xml, DanishXmlRuleDisambiguator(), "process-cached singleton")
}
