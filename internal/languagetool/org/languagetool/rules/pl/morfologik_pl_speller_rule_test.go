package pl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
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

func TestIsNotCompound_Prefix(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PolishSpellerDict, 1)
	sp.AddWord("postmodernistyczna")
	sp.AddWord("filozofów")
	r := NewMorfologikPolishSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// anty + postmodernistyczna (second longer than first)
	require.False(t, r.isNotCompound("antypostmodernistyczna"))
	// pre + moc: second "moc" length 3 == first "pre" length 3 → NOT accepted (Java: second.length() > first.length())
	// "premoc" should remain isNotCompound true (flaggable)
	require.True(t, r.isNotCompound("premoc"))
	// hiper + filozofów
	require.False(t, r.isNotCompound("hiperfilozofów"))
}

func TestIsNotCompound_AdjTagPOS(t *testing.T) {
	// NewMorfologikSpellerRule sets IsMisspelled that returns false when Words empty
	// (fail-closed) — prefix path then sees second as "accepted", but "zgniło" is not a prefix.
	r := NewMorfologikPolishSpellerRule()
	// without TagPOS: adj path skipped
	require.True(t, r.isNotCompound("zgniłożółty"))
	r.TagPOS = func(w string) []string {
		if w == "zgniło" {
			return []string{"adja"}
		}
		if w == "żółty" {
			return []string{"adj:sg:nom:m"}
		}
		return nil
	}
	require.False(t, r.isNotCompound("zgniłożółty"))
}

func TestAcceptTokenizedRemainder_NibyQuasi(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PolishSpellerDict, 1)
	sp.AddWord("artysta")
	sp.AddWord("opiekunem")
	sp.AddWord("Francuzem")
	r := NewMorfologikPolishSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.True(t, r.acceptTokenizedRemainder("Niby-artysta"))
	require.True(t, r.acceptTokenizedRemainder("quasi-opiekunem"))
	require.True(t, r.acceptTokenizedRemainder("niby-Francuzem"))
	require.False(t, r.acceptTokenizedRemainder("Niby-xyzzy"))
}

func TestMatch_LowerOnlySuggestion(t *testing.T) {
	// Java getRuleMatches: isMisspelled(original) true but !isMisspelled(lower) → only lower sug.
	// Map speller always case-folds; inject hooks that flag surface but accept lower.
	r := NewMorfologikPolishSpellerRule()
	r.IsMisspelled = func(w string) bool {
		return w != "zolw" // only exact lower accepted
	}
	sent := languagetool.AnalyzePlain("Zolw")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, m, 1)
	require.Equal(t, []string{"zolw"}, m[0].GetSuggestedReplacements())
}

func TestMatch_SuppressCompound(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PolishSpellerDict, 1)
	sp.AddWord("postmodernistyczna")
	// "antypostmodernistyczna" not in dict → base flags; isNotCompound suppresses
	r := NewMorfologikPolishSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("antypostmodernistyczna")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestMatch_NibyPrefix(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PolishSpellerDict, 1)
	sp.AddWord("artysta")
	r := NewMorfologikPolishSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("Niby-artysta")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestWirePolishSpellerTagPOS(t *testing.T) {
	r := NewMorfologikPolishSpellerRule()
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
