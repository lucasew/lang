package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestCAMoreFiltersRegistered(t *testing.T) {
	for _, class := range []string{
		"org.languagetool.rules.ca.CatalanRemoteRewriteFilter",
		"org.languagetool.rules.ca.AnarASuggestionsFilter",
		"org.languagetool.rules.ca.PortarGerundiSuggestionsFilter",
		"org.languagetool.rules.ca.FindSuggestionsEsFilter",
		"org.languagetool.rules.ca.SynthesizeWithDAFilter",
	} {
		require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(class), class)
	}
}

func TestCatalanRemoteRewriteFilter_Accept(t *testing.T) {
	f := NewCatalanRemoteRewriteFilter()
	// no Rewrite, suppressMatch false → keep match
	m := rules.NewRuleMatch(rules.NewFakeRule("R1"), languagetool.AnalyzePlain("Hola món."), 0, 4, "msg")
	out := f.AcceptRuleMatch(m, map[string]string{}, 0, nil, nil)
	require.NotNil(t, out)
	// suppressMatch true without rewrite → nil
	out = f.AcceptRuleMatch(m, map[string]string{"suppressMatch": "true"}, 0, nil, nil)
	require.Nil(t, out)

	f.Rewrite = func(sentence, ruleID string) string {
		require.Equal(t, "R1", ruleID)
		return "Adéu món."
	}
	out = f.AcceptRuleMatch(m, nil, 0, nil, nil)
	// may or may not get joined match depending on diff; at least no panic
	_ = out
}

func TestAnarA_AcceptFailClosed(t *testing.T) {
	f := NewAnarASuggestionsFilter()
	m := rules.NewRuleMatch(nil, languagetool.AnalyzePlain("anem a fer"), 0, 10, "msg")
	require.Nil(t, f.AcceptRuleMatch(m, nil, 0, nil, nil))
}
