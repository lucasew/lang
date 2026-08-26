package jekavian

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikJekavianSpellerRule_Paths(t *testing.T) {
	r := NewMorfologikJekavianSpellerRule()
	// Java MorfologikJekavianSpellerRule.getId / getFileName / path overrides
	require.Equal(t, "MORFOLOGIK_RULE_SR_JEKAVIAN", MorfologikJekavianSpellerRuleID)
	require.Equal(t, "/sr/dictionary/jekavian/serbian.dict", JekavianSpellerDict)
	require.Equal(t, MorfologikJekavianSpellerRuleID, r.GetID())
	require.Equal(t, JekavianSpellerDict, r.GetFileName())
	require.Equal(t, "/sr/dictionary/jekavian/ignored.txt", r.GetIgnoreFileName())
	require.Equal(t, "/sr/dictionary/jekavian/spelling.txt", r.GetSpellingFileName())
	require.Equal(t, "/sr/dictionary/jekavian/prohibit.txt", r.GetProhibitFileName())
	// Speller1 non-nil when official dict is discoverable (no invent).
	if morfologik.DiscoverLanguageDict(JekavianSpellerDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
	// Java does not override isLatinScript → default true → NonLatinScript false
	require.False(t, r.NonLatinScript)
}

func TestMorfologikJekavianSpellerRule_UserConfigAccepted(t *testing.T) {
	if morfologik.DiscoverLanguageDict(JekavianSpellerDict) == "" {
		t.Skip("serbian.dict (jekavian) not discoverable")
	}
	// Accepted words always go to wordsToBeIgnored (Java SpellingCheckRule ctor).
	// Use Latin surface: Java isLatinScript() default true ignores pure Cyrillic via pHasNoLetterLatin.
	uc := languagetool.NewUserConfigWithWords([]string{"xxxxyzzyqsrjkuser"}, nil)
	r := NewMorfologikJekavianSpellerRuleWithUser(uc)
	require.NotNil(t, r.Speller1)
	base := NewMorfologikJekavianSpellerRule()
	m, err := base.Match(languagetool.AnalyzePlain("xxxxyzzyqsrjkuser"))
	require.NoError(t, err)
	require.NotEmpty(t, m, "without user accepted words, invented surface is misspelled")
	m, err = r.Match(languagetool.AnalyzePlain("xxxxyzzyqsrjkuser"))
	require.NoError(t, err)
	require.Empty(t, m, "user accepted words must be ignored (Java wordsToBeIgnored)")
}
