package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	tagpl "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pl"
	"github.com/stretchr/testify/require"
)

func TestMorfologikPolishSpellerRule(t *testing.T) {
	r := NewMorfologikPolishSpellerRule()
	// Java MorfologikPolishSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_PL_PL", MorfologikPolishSpellerRuleID)
	require.Equal(t, "/pl/hunspell/pl_PL.dict", PolishSpellerDict)
	require.Equal(t, MorfologikPolishSpellerRuleID, r.GetID())
	require.Equal(t, PolishSpellerDict, r.GetFileName())
}

func TestPrunePolishSuggestions(t *testing.T) {
	// no space → keep
	got := prunePolishSuggestions([]string{"Clarke", "Clarkiem"})
	require.Equal(t, []string{"Clarke", "Clarkiem"}, got)
	// banned second token → drop
	got = prunePolishSuggestions([]string{"Clark em", "Clarke", "foo ze"})
	require.Equal(t, []string{"Clarke"}, got)
	// non-banned second → keep
	got = prunePolishSuggestions([]string{"Saint Exupery"})
	require.Equal(t, []string{"Saint Exupery"}, got)
}

func TestIsNotCompound_Prefix_RealDict(t *testing.T) {
	if morfologik.DiscoverLanguageDict(PolishSpellerDict) == "" {
		t.Skip("pl_PL.dict not discoverable")
	}
	r := NewMorfologikPolishSpellerRule()
	// anty + postmodernistyczna (second longer than first; second in dict)
	require.False(t, r.isNotCompound("antypostmodernistyczna"))
	// pre + moc: second "moc" length 3 == first "pre" length 3 → NOT accepted
	// "premoc" should remain isNotCompound true (flaggable)
	require.True(t, r.isNotCompound("premoc"))
	// hiper + filozofów — filozofów may not be in speller dict alone; skip if not accepted
	// Java accepts via second not misspelled. If dict has it, compound suppresses.
	if !r.wordIsMisspelled("filozofów") {
		require.False(t, r.isNotCompound("hiperfilozofów"))
	}
}

func TestIsNotCompound_AdjTagPOS_RealTagger(t *testing.T) {
	if morfologik.DiscoverLanguageDict(PolishSpellerDict) == "" {
		t.Skip("pl_PL.dict not discoverable")
	}
	if tagpl.DiscoverPolishPOSDict() == "" {
		t.Skip("polish.dict not discoverable")
	}
	r := NewMorfologikPolishSpellerRule()
	// NewMorfologikPolishSpellerRule wires TagPOS from PolishTagger when present.
	require.NotNil(t, r.TagPOS, "TagPOS must wire from PolishTagger when polish.dict present")
	// zgniło (adja) + żółty (adj:) — isNotCompound returns false
	require.False(t, r.isNotCompound("zgniłożółty"))
	// trzynasto (adja/num:comp) + bitowe (adj:)
	require.False(t, r.isNotCompound("trzynastobitowe"))
}

func TestMatch_NibyPrefix_RealDict(t *testing.T) {
	if morfologik.DiscoverLanguageDict(PolishSpellerDict) == "" {
		t.Skip("pl_PL.dict not discoverable")
	}
	r := NewMorfologikPolishSpellerRule()
	// artysta is in pl_PL.dict → remainder accepted after tokenizingPattern split
	sent := languagetool.AnalyzePlain("Niby-artysta")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestMatch_SuppressCompound_RealDict(t *testing.T) {
	if morfologik.DiscoverLanguageDict(PolishSpellerDict) == "" {
		t.Skip("pl_PL.dict not discoverable")
	}
	r := NewMorfologikPolishSpellerRule()
	sent := languagetool.AnalyzePlain("antypostmodernistyczna")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestWirePolishSpellerTagPOS(t *testing.T) {
	r := NewMorfologikPolishSpellerRule()
	// clear TagPOS to test inject path
	r.TagPOS = nil
	require.True(t, r.isNotCompound("zgniłożółty"))
	WirePolishSpellerTagPOS(r, func(token string) []languagetool.TokenTag {
		if token == "zgniło" {
			return []languagetool.TokenTag{{POS: "adja"}}
		}
		if token == "żółty" {
			return []languagetool.TokenTag{{POS: "adj:sg:nom:m"}}
		}
		return nil
	})
	require.False(t, r.isNotCompound("zgniłożółty"))
}
