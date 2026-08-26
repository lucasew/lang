package sl

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikSlovenianSpellerRule(t *testing.T) {
	r := NewMorfologikSlovenianSpellerRule()
	// Java MorfologikSlovenianSpellerRule.getId / RESOURCE_FILENAME
	require.Equal(t, "MORFOLOGIK_RULE_SL_SI", MorfologikSlovenianSpellerRuleID)
	require.Equal(t, "/sl/hunspell/sl_SI.dict", SlovenianSpellerDict)
	require.Equal(t, MorfologikSlovenianSpellerRuleID, r.GetID())
	require.Equal(t, SlovenianSpellerDict, r.GetFileName())
	// Speller1 non-nil when official dict is discoverable (no invent).
	if morfologik.DiscoverLanguageDict(SlovenianSpellerDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
}

func TestMorfologikSlovenianSpellerRule_UserConfigAccepted(t *testing.T) {
	if morfologik.DiscoverLanguageDict(SlovenianSpellerDict) == "" {
		t.Skip("sl_SI.dict not discoverable")
	}
	// Accepted words always go to wordsToBeIgnored (Java SpellingCheckRule ctor).
	uc := languagetool.NewUserConfigWithWords([]string{"xxxxyzzyqsluser"}, nil)
	r := NewMorfologikSlovenianSpellerRuleWithUser(uc)
	require.NotNil(t, r.Speller1)
	// Pure super: misspelled surface without user config is flagged; with ignore, not.
	base := NewMorfologikSlovenianSpellerRule()
	m, err := base.Match(languagetool.AnalyzePlain("xxxxyzzyqsluser"))
	require.NoError(t, err)
	require.NotEmpty(t, m, "without user accepted words, invented surface is misspelled")
	m, err = r.Match(languagetool.AnalyzePlain("xxxxyzzyqsluser"))
	require.NoError(t, err)
	require.Empty(t, m, "user accepted words must be ignored (Java wordsToBeIgnored)")
}
