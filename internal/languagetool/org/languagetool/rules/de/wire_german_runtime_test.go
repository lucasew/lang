package de

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDiscoverGermanHunspellDict(t *testing.T) {
	// Present in inspiration tree when checkout includes DE module resources.
	p := DiscoverGermanHunspellDict("DE")
	if p == "" {
		t.Skip("no de_DE.dict in workspace")
	}
	require.Contains(t, p, "de_DE.dict")
	// AT/CH fall back if needed
	_ = DiscoverGermanHunspellDict("AT")
	_ = DiscoverGermanHunspellDict("CH")
}

func TestDiscoverGermanGrammarXML(t *testing.T) {
	p := DiscoverGermanGrammarXML()
	if p == "" {
		t.Skip("no grammar.xml in workspace")
	}
	require.Contains(t, p, "grammar.xml")
	st := DiscoverGermanStyleXML()
	if st != "" {
		require.Contains(t, st, "style.xml")
	}
}

func TestDiscoverGermanRemoteRuleFiltersXML(t *testing.T) {
	p := DiscoverGermanRemoteRuleFiltersXML()
	if p == "" {
		t.Skip("no remote-rule-filters.xml in workspace")
	}
	require.Contains(t, p, "remote-rule-filters.xml")
}

func TestDiscoverGermanDEATGrammarXML(t *testing.T) {
	// Java GermanyGerman/AustrianGerman/NonSwissGerman getRuleFileNames.
	p := DiscoverGermanDEATGrammarXML()
	if p == "" {
		t.Skip("no de-DE-AT/grammar.xml in workspace")
	}
	require.Contains(t, p, "de-DE-AT")
	require.Contains(t, p, "grammar.xml")
}

func TestDiscoverGermanCHGrammarXML(t *testing.T) {
	// Java Language.getRuleFileNames: de/de-CH/grammar.xml for SwissGerman.
	p := DiscoverGermanCHGrammarXML()
	if p == "" {
		t.Skip("no de-CH/grammar.xml in workspace")
	}
	require.Contains(t, p, "de-CH")
	require.Contains(t, p, "grammar.xml")
}

func TestWireGermanRuntimeResources_NoPanic(t *testing.T) {
	require.NotPanics(t, func() {
		WireGermanRuntimeResources("DE")
	})
}

func TestWireGermanUpstreamGrammar_Gated(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	// Without env: no invent load of multi-MB grammar in unit tests.
	prev := os.Getenv("LANG_USE_UPSTREAM_GRAMMAR")
	_ = os.Unsetenv("LANG_USE_UPSTREAM_GRAMMAR")
	t.Cleanup(func() {
		if prev != "" {
			_ = os.Setenv("LANG_USE_UPSTREAM_GRAMMAR", prev)
		}
	})
	before := len(lt.GetAllRegisteredRuleIDs())
	WireGermanUpstreamGrammar(lt)
	after := len(lt.GetAllRegisteredRuleIDs())
	require.Equal(t, before, after, "must not load grammar without LANG_USE_UPSTREAM_GRAMMAR=1")
}
