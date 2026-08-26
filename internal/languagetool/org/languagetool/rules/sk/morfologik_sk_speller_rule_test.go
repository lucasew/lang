package sk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikSlovakSpellerRule(t *testing.T) {
	r := NewMorfologikSlovakSpellerRule()
	// Java MorfologikSlovakSpellerRule.getId / RESOURCE_FILENAME
	require.Equal(t, "MORFOLOGIK_RULE_SK_SK", MorfologikSlovakSpellerRuleID)
	require.Equal(t, "/sk/hunspell/sk_SK.dict", SlovakSpellerDict)
	require.Equal(t, MorfologikSlovakSpellerRuleID, r.GetID())
	require.Equal(t, SlovakSpellerDict, r.GetFileName())
	// Speller1 non-nil when official dict is discoverable (no invent).
	if morfologik.DiscoverLanguageDict(SlovakSpellerDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
}

func TestMorfologikSlovakSpellerRule_UserConfigAccepted(t *testing.T) {
	if morfologik.DiscoverLanguageDict(SlovakSpellerDict) == "" {
		t.Skip("sk_SK.dict not discoverable")
	}
	// Accepted words always go to wordsToBeIgnored (Java SpellingCheckRule ctor).
	uc := languagetool.NewUserConfigWithWords([]string{"xxxxyzzyqskuser"}, nil)
	r := NewMorfologikSlovakSpellerRuleWithUser(uc)
	require.NotNil(t, r.Speller1)
	// Pure super: misspelled surface without user config is flagged; with ignore, not.
	base := NewMorfologikSlovakSpellerRule()
	m, err := base.Match(languagetool.AnalyzePlain("xxxxyzzyqskuser"))
	require.NoError(t, err)
	require.NotEmpty(t, m, "without user accepted words, invented surface is misspelled")
	m, err = r.Match(languagetool.AnalyzePlain("xxxxyzzyqskuser"))
	require.NoError(t, err)
	require.Empty(t, m, "user accepted words must be ignored (Java wordsToBeIgnored)")
}
