package pl

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Outcome twins for PolishHybridDisambiguator.disambiguator field:
// new XmlRuleDisambiguator(new Polish()) → useGlobalDisambiguation=false
// over official resource/pl/disambiguation.xml (Java /pl/disambiguation.xml).

func TestPolishXmlRuleDisambiguator_ProcessCachedOfficial(t *testing.T) {
	require.NotEmpty(t, DiscoverPolishDisambiguationXML(),
		"official pl/disambiguation.xml must be discoverable (inspiration pl module)")
	p := DiscoverPolishDisambiguationXML()
	require.Contains(t, p, "disambiguation.xml")
	require.Contains(t, p, filepath.Join("resource", "pl"))

	a := PolishXmlRuleDisambiguator()
	b := PolishXmlRuleDisambiguator()
	require.NotNil(t, a, "XmlRuleDisambiguator(Polish) must load official disambiguation.xml")
	require.Same(t, a, b, "process-cached singleton")
	// Official PL pack is large; language-only (useGlobal=false).
	require.GreaterOrEqual(t, len(a.Rules), 200,
		"Polish XmlRuleDisambiguator must load official pl/disambiguation.xml rules")

	// Wired on NewPolishHybridDisambiguator (Java final field disambiguator).
	d := NewPolishHybridDisambiguator()
	require.NotNil(t, d.Rules)
	require.Same(t, a, d.Rules)
}
