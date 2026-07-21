package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of Java MorfologikMultiSpeller user-dict first + weighted merge:
// user accepted words form a separate speller; getSuggestionsFromUserDicts /
// getSuggestionsFromDefaultDicts split; calcSpellerSuggestions concatenates
// user before default when word.length() > 4.

func TestMultiSpeller_UserDictSuggestions(t *testing.T) {
	main := NewMorfologikSpeller("/xx/spelling/test.dict", 1)
	main.AddWord("defaultfix")
	main.Suggestions["defalt"] = []string{"defaultfix"}

	user := NewMorfologikSpeller("/xx/spelling/user", 1)
	user.AddWord("userform")
	user.Suggestions["usreform"] = []string{"userform"}
	user.Suggestions["defalt"] = []string{"userform"} // personal correction for same typo

	m := &MorfologikMultiSpeller{
		Spellers:            []*MorfologikSpeller{user, main},
		UserDictSpellers:    []*MorfologikSpeller{user},
		DefaultDictSpellers: []*MorfologikSpeller{main},
		BinaryDictPath:      "/xx/spelling/test.dict",
	}

	// Membership: either dict accepts
	require.False(t, m.IsMisspelled("userform"))
	require.False(t, m.IsMisspelled("defaultfix"))
	require.True(t, m.IsMisspelled("usreform"))

	// Split APIs
	require.Equal(t, []string{"userform"}, m.GetSuggestionsFromUserDicts("usreform"))
	require.Empty(t, m.GetSuggestionsFromDefaultDicts("usreform"))
	require.Equal(t, []string{"userform"}, m.GetSuggestionsFromUserDicts("defalt"))
	require.Equal(t, []string{"defaultfix"}, m.GetSuggestionsFromDefaultDicts("defalt"))

	// Full getSuggestions merges + weight-sorts; both present for "defalt"
	all := m.GetSuggestions("defalt")
	require.Contains(t, all, "userform")
	require.Contains(t, all, "defaultfix")
}

func TestMorfologikSpellerRule_UserDictOrdering(t *testing.T) {
	main := NewMorfologikSpeller("/xx/spelling/test.dict", 1)
	main.AddWord("software")
	main.Suggestions["sofware"] = []string{"software"} // length 7 > 4

	user := NewMorfologikSpeller("/xx/spelling/user", 1)
	user.AddWord("sofwarex") // not used
	user.AddWord("mytool")
	user.Suggestions["sofware"] = []string{"mytool"}

	m := &MorfologikMultiSpeller{
		Spellers:            []*MorfologikSpeller{user, main},
		UserDictSpellers:    []*MorfologikSpeller{user},
		DefaultDictSpellers: []*MorfologikSpeller{main},
	}
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/xx/spelling/test.dict", main)
	r.Multi = m
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}

	// word.length()>4 → user suggestions before default
	sugs := r.collectSuggestions("sofware")
	require.GreaterOrEqual(t, len(sugs), 2)
	require.Equal(t, "mytool", sugs[0], "user dict first for long words; got %v", sugs)
	require.Contains(t, sugs, "software")

	// short word length <= 4 → default before user
	main.Suggestions["abx"] = []string{"abc"}
	user.Suggestions["abx"] = []string{"abz"}
	main.AddWord("abc")
	user.AddWord("abz")
	sugsShort := r.collectSuggestions("abx")
	require.GreaterOrEqual(t, len(sugsShort), 2)
	require.Equal(t, "abc", sugsShort[0], "default first for short words; got %v", sugsShort)
	require.Contains(t, sugsShort, "abz")

	// End-to-end Match surfaces ordered suggestions
	ms, err := r.Match(languagetool.AnalyzePlain("sofware"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	got := ms[0].GetSuggestedReplacements()
	require.NotEmpty(t, got)
	require.Equal(t, "mytool", got[0], "Match suggestions: %v", got)
}

func TestOpenMultiSpellerWithUser_AcceptsAndSuggests(t *testing.T) {
	if DiscoverLanguageDict("/en/hunspell/en_US.dict") == "" {
		t.Skip("en_US.dict not in tree")
	}
	m := OpenMultiSpellerFromClasspathWithUser(
		"/en/hunspell/en_US.dict",
		nil, "", 1, nil,
		[]string{"Zyxword", "CorpName"},
	)
	require.NotNil(t, m)
	require.Len(t, m.UserDictSpellers, 1)
	require.False(t, m.IsMisspelled("Zyxword"))
	require.False(t, m.IsMisspelled("software")) // binary
	require.True(t, m.IsMisspelled("Zyxwordx"))
	// edit-1 of Zyxword
	userSugs := m.GetSuggestionsFromUserDicts("Zyxwor")
	require.Contains(t, userSugs, "Zyxword")
	// not from default
	require.NotContains(t, m.GetSuggestionsFromDefaultDicts("Zyxwor"), "Zyxword")
}

func TestUTF16Len(t *testing.T) {
	require.Equal(t, 5, UTF16Len("hello"))
	require.Equal(t, 3, UTF16Len("abx"))
	// BMP letter still one UTF-16 unit
	require.Equal(t, 1, UTF16Len("é"))
}
