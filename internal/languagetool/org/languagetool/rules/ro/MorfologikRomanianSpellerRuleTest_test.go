package ro

// Twin surface for MorfologikRomanianSpellerRule (Java has no dedicated *RuleTest class).
// Official resource: /ro/hunspell/ro_RO.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.
// Latin script default (Java does not override isLatinScript).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireRomanianSpellerDict skips when official ro_RO.dict is not discoverable (no invent).
func requireRomanianSpellerDict(t *testing.T) *MorfologikRomanianSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(RomanianSpellerDict) == "" {
		t.Skip("ro_RO.dict not discoverable (inspiration/.../ro/hunspell)")
	}
	r := NewMorfologikRomanianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// Metadata + surface: getId / getFileName / Latin default.
func TestMorfologikRomanianSpellerRule_Metadata(t *testing.T) {
	r := NewMorfologikRomanianSpellerRule()
	require.Equal(t, MorfologikRomanianSpellerRuleID, r.GetID())
	require.Equal(t, RomanianSpellerDict, r.GetFileName())
	require.Equal(t, "MORFOLOGIK_RULE_RO_RO", r.GetID())
	require.Equal(t, "/ro/hunspell/ro_RO.dict", r.GetFileName())
	// Java pure super: isLatinScript default true → NonLatinScript false.
	if r.SpellingCheckRule != nil {
		require.False(t, r.NonLatinScript)
	}
	if morfologik.DiscoverLanguageDict(RomanianSpellerDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
}

// Real-dict Match matrix: known good RO → 0; deliberate misspelling → ≥1.
func TestMorfologikRomanianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireRomanianSpellerDict(t)

	matchN := func(s string) int {
		t.Helper()
		ms, err := r.Match(languagetool.AnalyzePlain(s))
		require.NoError(t, err)
		return len(ms)
	}

	// correct words (official ro_RO.dict / common RO closed class)
	require.Equal(t, 0, matchN("este"))
	require.Equal(t, 0, matchN("pentru"))
	require.Equal(t, 0, matchN("din"))
	require.Equal(t, 0, matchN("limba"))
	require.Equal(t, 0, matchN("România"))
	require.Equal(t, 0, matchN("cuvânt"))
	require.Equal(t, 0, matchN("și"))

	// deliberate misspellings
	require.GreaterOrEqual(t, matchN("xxxxyzzyq"), 1)
	require.GreaterOrEqual(t, matchN("estee"), 1)
	require.GreaterOrEqual(t, matchN("romaniia"), 1)

	// digits / punctuation not flagged (no letters → skipped by parent Match)
	require.Equal(t, 0, matchN("123"))
	require.Equal(t, 0, matchN(","))
}

func TestMorfologikRomanianSpellerRule_UserConfigAccepted(t *testing.T) {
	if morfologik.DiscoverLanguageDict(RomanianSpellerDict) == "" {
		t.Skip("ro_RO.dict not discoverable")
	}
	// Accepted words always go to wordsToBeIgnored (Java SpellingCheckRule ctor).
	uc := languagetool.NewUserConfigWithWords([]string{"xxxxyzzyqrouser"}, nil)
	r := NewMorfologikRomanianSpellerRuleWithUser(uc)
	require.NotNil(t, r.Speller1)
	// Pure super: misspelled surface without user config is flagged; with ignore, not.
	base := NewMorfologikRomanianSpellerRule()
	m, err := base.Match(languagetool.AnalyzePlain("xxxxyzzyqrouser"))
	require.NoError(t, err)
	require.NotEmpty(t, m, "without user accepted words, invented surface is misspelled")
	m, err = r.Match(languagetool.AnalyzePlain("xxxxyzzyqrouser"))
	require.NoError(t, err)
	require.Empty(t, m, "user accepted words must be ignored (Java wordsToBeIgnored)")
}
