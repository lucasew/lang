package it

// Constructor / load twin for ItalianRuleDisambiguator (outcome myAssert twins live under rules/it).

import (
	"testing"

	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/stretchr/testify/require"
)

func TestItalianRuleDisambiguator_LoadsOfficialXML(t *testing.T) {
	if DiscoverItalianDisambiguationXML() == "" {
		t.Skip("it/disambiguation.xml not in tree")
	}
	d := NewItalianRuleDisambiguator()
	require.NotNil(t, d)
	require.NotNil(t, d.Rules, "Java eagerly constructs XmlRuleDisambiguator(Italian)")
	xml, ok := d.Rules.(*disambigrules.XmlRuleDisambiguator)
	require.True(t, ok)
	require.NotEmpty(t, xml.Rules)
	require.NotNil(t, xml.UnifierConfig)
	// nil input: Go returns nil (disambiguator stage guard)
	require.Nil(t, d.Disambiguate(nil))
}
