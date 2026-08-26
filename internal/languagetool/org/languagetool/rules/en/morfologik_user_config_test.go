package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Twin of Java MorfologikMultiSpeller + SpellingCheckRule UserConfig:
// - accepted words always ignored (wordsToBeIgnored)
// - user-dict FSA / suggestions only when premiumUid != null

func TestAmericanSpeller_UserConfigFree_IgnoreOnly(t *testing.T) {
	if morfologik.DiscoverLanguageDict(AmericanSpellerDict) == "" {
		t.Skip("en_US.dict not in tree")
	}
	uc := languagetool.NewUserConfigWithWords([]string{"Zyxuserword"}, nil)
	// free: PremiumUID nil
	require.Nil(t, uc.GetPremiumUid())
	r := NewMorfologikAmericanSpellerRuleWithUser(uc)
	require.NotNil(t, r.Multi)
	require.Empty(t, r.Multi.UserDictSpellers, "free account: no user FSA")
	// ignored via SpellingCheckRule
	require.True(t, r.AcceptWord("Zyxuserword"))
	ms, err := r.Match(analyzeEN("Zyxuserword is fine."))
	require.NoError(t, err)
	// "is"/"fine" may flag if not ignored length/lists; Zyxuserword must not
	for _, m := range ms {
		covered := "Zyxuserword is fine."[m.GetFromPos():m.GetToPos()]
		require.NotEqual(t, "Zyxuserword", covered)
	}
}

func TestAmericanSpeller_UserConfigPremium_UserDictSuggests(t *testing.T) {
	if morfologik.DiscoverLanguageDict(AmericanSpellerDict) == "" {
		t.Skip("en_US.dict not in tree")
	}
	uid := int64(42)
	uc := languagetool.NewUserConfigWithWords([]string{"CorpName"}, nil)
	uc.PremiumUID = &uid
	r := NewMorfologikAmericanSpellerRuleWithUser(uc)
	require.NotNil(t, r.Multi)
	require.Len(t, r.Multi.UserDictSpellers, 1)
	require.False(t, r.Multi.IsMisspelled("CorpName"))
	// edit-1 suggestion from user dict
	userSugs := r.Multi.GetSuggestionsFromUserDicts("CorpNam")
	require.Contains(t, userSugs, "CorpName")
	// long word: user suggestions before default (Match surfaces ordered sugs)
	ms, err := r.Match(analyzeEN("CorpNam"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	sugs := ms[0].GetSuggestedReplacements()
	require.NotEmpty(t, sugs)
	require.Equal(t, "CorpName", sugs[0], "user dict first for len>4; got %v", sugs)
}

func TestUserDictWordsForMulti_Gate(t *testing.T) {
	uid := int64(1)
	require.Nil(t, morfologik.UserDictWordsForMulti([]string{"a"}, nil))
	require.Nil(t, morfologik.UserDictWordsForMulti(nil, &uid))
	require.Equal(t, []string{"a", "b"}, morfologik.UserDictWordsForMulti([]string{"a", "b"}, &uid))
}
