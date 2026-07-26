package el

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikGreekSpellerRule(t *testing.T) {
	r := NewMorfologikGreekSpellerRule()
	// Java MorfologikGreekSpellerRule.getId / RESOURCE_FILENAME
	require.Equal(t, "MORFOLOGIK_RULE_EL_GR", MorfologikGreekSpellerRuleID)
	require.Equal(t, "/el/hunspell/el_GR.dict", GreekSpellerDict)
	require.Equal(t, MorfologikGreekSpellerRuleID, r.GetID())
	require.Equal(t, GreekSpellerDict, r.GetFileName())
	// Java isLatinScript() = false → NonLatinScript
	require.False(t, r.IsLatinScript())
	require.True(t, r.NonLatinScript)
	// Speller1 non-nil when official dict is discoverable (no invent).
	if morfologik.DiscoverLanguageDict(GreekSpellerDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
}

func TestMorfologikGreekSpellerRule_UserConfigAccepted(t *testing.T) {
	if morfologik.DiscoverLanguageDict(GreekSpellerDict) == "" {
		t.Skip("el_GR.dict not discoverable")
	}
	// Accepted words always go to wordsToBeIgnored (Java SpellingCheckRule ctor).
	uc := languagetool.NewUserConfigWithWords([]string{"xxxxyzzyqeluser"}, nil)
	r := NewMorfologikGreekSpellerRuleWithUser(uc)
	require.NotNil(t, r.Speller1)
	// Pure super: misspelled surface without user config is flagged; with ignore, not.
	base := NewMorfologikGreekSpellerRule()
	m, err := base.Match(languagetool.AnalyzePlain("xxxxyzzyqeluser"))
	require.NoError(t, err)
	require.NotEmpty(t, m, "without user accepted words, invented surface is misspelled")
	m, err = r.Match(languagetool.AnalyzePlain("xxxxyzzyqeluser"))
	require.NoError(t, err)
	require.Empty(t, m, "user accepted words must be ignored (Java wordsToBeIgnored)")
}
