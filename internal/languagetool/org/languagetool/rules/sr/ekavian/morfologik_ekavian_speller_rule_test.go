package ekavian

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikEkavianSpellerRule_Paths(t *testing.T) {
	r := NewMorfologikEkavianSpellerRule()
	// Java MorfologikEkavianSpellerRule.getId / getFileName / path overrides
	require.Equal(t, "MORFOLOGIK_RULE_SR_EKAVIAN", MorfologikEkavianSpellerRuleID)
	require.Equal(t, "/sr/dictionary/ekavian/serbian.dict", EkavianSpellerDict)
	require.Equal(t, MorfologikEkavianSpellerRuleID, r.GetID())
	require.Equal(t, EkavianSpellerDict, r.GetFileName())
	require.Equal(t, "/sr/dictionary/ekavian/ignored.txt", r.GetIgnoreFileName())
	require.Equal(t, "/sr/dictionary/ekavian/spelling.txt", r.GetSpellingFileName())
	require.Equal(t, "/sr/dictionary/ekavian/prohibit.txt", r.GetProhibitFileName())
	// Speller1 non-nil when official dict is discoverable (no invent).
	if morfologik.DiscoverLanguageDict(EkavianSpellerDict) != "" {
		require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	}
	// Java does not override isLatinScript → default true → NonLatinScript false
	require.False(t, r.NonLatinScript)
}

func TestMorfologikEkavianSpellerRule_ExamplePair(t *testing.T) {
	r := NewMorfologikEkavianSpellerRule()
	require.NotEmpty(t, r.GetIncorrectExamples())
	require.NotEmpty(t, r.GetCorrectExamples())
	inc := r.GetIncorrectExamples()[0]
	// Java Example.wrong / fixed surfaces
	require.Contains(t, inc.GetExample(), "бткие")
	require.Equal(t, []string{"битке"}, inc.GetCorrections())
	cor := r.GetCorrectExamples()[0]
	require.Contains(t, cor.GetExample(), "битке")
}

func TestMorfologikEkavianSpellerRule_UserConfigAccepted(t *testing.T) {
	if morfologik.DiscoverLanguageDict(EkavianSpellerDict) == "" {
		t.Skip("serbian.dict (ekavian) not discoverable")
	}
	// Accepted words always go to wordsToBeIgnored (Java SpellingCheckRule ctor).
	// Use Latin surface: Java isLatinScript() default true ignores pure Cyrillic via pHasNoLetterLatin.
	uc := languagetool.NewUserConfigWithWords([]string{"xxxxyzzyqsrekuser"}, nil)
	r := NewMorfologikEkavianSpellerRuleWithUser(uc)
	require.NotNil(t, r.Speller1)
	base := NewMorfologikEkavianSpellerRule()
	m, err := base.Match(languagetool.AnalyzePlain("xxxxyzzyqsrekuser"))
	require.NoError(t, err)
	require.NotEmpty(t, m, "without user accepted words, invented surface is misspelled")
	m, err = r.Match(languagetool.AnalyzePlain("xxxxyzzyqsrekuser"))
	require.NoError(t, err)
	require.Empty(t, m, "user accepted words must be ignored (Java wordsToBeIgnored)")
}
