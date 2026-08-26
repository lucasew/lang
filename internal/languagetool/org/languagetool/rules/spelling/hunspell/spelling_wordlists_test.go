package hunspell

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

func TestDiscoverLangHunspellWordList_DanishIgnore(t *testing.T) {
	p := DiscoverLangHunspellWordList("da", "ignore.txt")
	if p == "" {
		t.Skip("da ignore.txt not in tree")
	}
	require.Contains(t, p, "ignore.txt")
	words, err := LoadSpellingWordListFile(p)
	require.NoError(t, err)
	require.NotEmpty(t, words)
	// sample from da ignore.txt (see module file)
	require.Contains(t, words, "kr")
}

func TestApplyDefaultSpellingWordLists_IgnoreAndProhibit(t *testing.T) {
	if DiscoverLangHunspellWordList("da", "ignore.txt") == "" {
		t.Skip("da ignore.txt not in tree")
	}
	r := spelling.NewSpellingCheckRule("HUNSPELL_RULE", "spell", "da")
	// pretend everything is misspelled unless ignore/prohibit path says otherwise
	r.IsMisspelled = func(string) bool { return true }
	ApplyDefaultSpellingWordLists(r)
	require.True(t, r.AcceptWord("kr"), "ignore.txt word must be accepted")
	// garbage still misspelled
	require.False(t, r.AcceptWord("xyzzyqqqnotaword"))
}

func TestApplyDefaultSpellingWordLists_ArabicProhibit(t *testing.T) {
	if DiscoverLangHunspellWordList("ar", "prohibit.txt") == "" {
		t.Skip("ar prohibit.txt not in tree")
	}
	r := spelling.NewSpellingCheckRule("HUNSPELL_RULE_AR", "spell", "ar")
	// dict would accept everything
	r.IsMisspelled = func(string) bool { return false }
	ApplyDefaultSpellingWordLists(r)
	// if prohibit has real entries, they must not be accepted
	if len(r.ProhibitedWords) == 0 {
		t.Skip("ar prohibit.txt has no active (uncommented) entries")
	}
	for w := range r.ProhibitedWords {
		require.True(t, r.IsProhibited(w))
		require.False(t, r.AcceptWord(w), w)
		break
	}
}
