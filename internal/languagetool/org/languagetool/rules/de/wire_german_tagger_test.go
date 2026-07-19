package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWireGermanTagWord_NoPanic(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	ok := WireGermanTagWord(lt)
	if !ok {
		// manual-only or dict missing still may open from added.txt
		if DiscoveredGermanTagger() == nil {
			t.Skip("no German tagger resources")
		}
	}
	require.NotNil(t, lt.TagWord)
	// unknown → empty (no invent)
	require.Empty(t, lt.TagWord("xyzzyqnotaword123"))
}

func TestGermanTagWord_NilTagger(t *testing.T) {
	require.Nil(t, GermanTagWord(nil))
	require.False(t, WireGermanTagWord(nil))
}

func TestRegisterCore_WiresTagWordWhenResources(t *testing.T) {
	if DiscoveredGermanTagger() == nil {
		t.Skip("no German tagger resources")
	}
	lt := languagetool.NewJLanguageTool("de")
	RegisterCoreGermanRules(lt)
	require.NotNil(t, lt.TagWord, "Java createDefaultTagger must be wired")
	// Analyze path should use TagWord (may still leave unknown untagged)
	sents := lt.Analyze("Das Haus.")
	require.NotEmpty(t, sents)
}

func TestWireGermanTagWordFor_CH_NoPanic(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-CH")
	// May return false without dict — must not panic; when ok TagWord is set.
	ok := WireGermanTagWordFor(lt, "CH")
	if ok {
		require.NotNil(t, lt.TagWord)
	}
}
