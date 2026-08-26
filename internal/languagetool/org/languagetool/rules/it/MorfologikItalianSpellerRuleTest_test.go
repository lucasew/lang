package it

// Twin of org.languagetool.rules.it.MorfologikItalianSpellerRuleTest
// Official resource: /it/hunspell/it_IT.dict (+ spelling.txt via InitSpellersFromGetters).
// No map-inject / ClearMultiSpellers invent — real binary dict only.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// requireItalianSpellerDict skips when official it_IT.dict is not discoverable (no invent).
func requireItalianSpellerDict(t *testing.T) *MorfologikItalianSpellerRule {
	t.Helper()
	if morfologik.DiscoverLanguageDict(ItalianSpellerDict) == "" {
		t.Skip("it_IT.dict not discoverable (inspiration/.../it/hunspell)")
	}
	r := NewMorfologikItalianSpellerRule()
	require.NotNil(t, r)
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	return r
}

// Port of MorfologikItalianSpellerRuleTest.testMorfologikSpeller
func TestMorfologikItalianSpellerRule_MorfologikSpeller(t *testing.T) {
	r := requireItalianSpellerDict(t)
	require.Equal(t, MorfologikItalianSpellerRuleID, r.GetID())
	require.Equal(t, ItalianSpellerDict, r.GetFileName())

	// Java: matches = rule.match(lt.getAnalyzedSentence("esmpio"));
	// assertEquals(1, matches.length);
	// assertEquals("[empio, esempio, espio, espiò]", matches[0].getSuggestedReplacements().toString());
	ms, err := r.Match(languagetool.AnalyzePlain("esmpio"))
	require.NoError(t, err)
	require.Len(t, ms, 1)
	require.Equal(t, []string{"empio", "esempio", "espio", "espiò"}, ms[0].GetSuggestedReplacements(),
		"Java MorfologikItalianSpellerRuleTest exact suggestion list (order matters)")
}

func TestMorfologikItalianSpellerRule_UserConfigAccepted(t *testing.T) {
	if morfologik.DiscoverLanguageDict(ItalianSpellerDict) == "" {
		t.Skip("it_IT.dict not discoverable")
	}
	// Accepted words always go to wordsToBeIgnored (Java SpellingCheckRule ctor).
	uc := languagetool.NewUserConfigWithWords([]string{"xxxxyzzyqituser"}, nil)
	r := NewMorfologikItalianSpellerRuleWithUser(uc)
	require.NotNil(t, r.Speller1)
	base := NewMorfologikItalianSpellerRule()
	m, err := base.Match(languagetool.AnalyzePlain("xxxxyzzyqituser"))
	require.NoError(t, err)
	require.NotEmpty(t, m, "without user accepted words, invented surface is misspelled")
	m, err = r.Match(languagetool.AnalyzePlain("xxxxyzzyqituser"))
	require.NoError(t, err)
	require.Empty(t, m, "user accepted words must be ignored (Java wordsToBeIgnored)")
}
