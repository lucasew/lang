package sk

// Twin of org.languagetool.rules.sk.LanguageSpecificSpellcheckerTest
//   → SpellcheckerTest.runLanguageSpecificTest()
//
// Java (SpellcheckerTest.runLanguageSpecificTest) for each Language on classpath:
//   prohibitedWords = CachingWordListLoader.loadWords(shortCode + "/hunspell/prohibit.txt")
//   spellingWords   = CachingWordListLoader.loadWords(shortCode + "/hunspell/spelling.txt")
//   intersection(prohibited distinct ∩ spelling) must be empty
//
// SK module test only invokes runLanguageSpecificTest(). SK-visible paths:
//   sk/hunspell/spelling.txt (present; comment-only upstream → empty word list)
//   sk/hunspell/prohibit.txt (absent → empty list, Java resourceExists false)
// Missing resources → empty (no invent). Intersection for SK must be empty.

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of LanguageSpecificSpellcheckerTest.testRules → runLanguageSpecificTest (SK sector).
func TestLanguageSpecificSpellchecker_Rules(t *testing.T) {
	// Official SK spelling list path (Java DataBroker sk/hunspell/spelling.txt).
	skSpellingPath := spelling.DiscoverSpellingResource("sk/hunspell/spelling.txt")
	if skSpellingPath == "" {
		t.Skip("official sk/hunspell/spelling.txt not discoverable — honest skip, no invent")
	}

	// SK-visible outcomes of SpellcheckerTest for short code "sk".
	// Java CachingWordListLoader: missing prohibit.txt → empty list (not invent).
	prohibitedWords := loadHunspellWordList(t, "sk", "prohibit.txt")
	spellingWords := loadHunspellWordList(t, "sk", "spelling.txt")
	t.Logf("Testing sk: %d words from prohibit.txt and %d words from spelling.txt",
		len(prohibitedWords), len(spellingWords))
	require.Empty(t, wordListIntersection(prohibitedWords, spellingWords),
		"sk prohibit.txt ∩ spelling.txt must be empty")

	// Speller rule surface still loads (LanguageSpecificSpellcheckerTest lives next to the rule).
	r := NewMorfologikSlovakSpellerRule()
	require.Equal(t, MorfologikSlovakSpellerRuleID, r.GetID())
	require.Equal(t, SlovakSpellerDict, r.GetFileName())
	if morfologik.DiscoverLanguageDict(SlovakSpellerDict) != "" {
		require.NotNil(t, r.Speller1)
	}

	// Analyze surface (JLanguageTool sk).
	require.NotEmpty(t, languagetool.NewJLanguageTool("sk").Analyze("test"))
}

// loadHunspellWordList ports CachingWordListLoader.loadWords(short + "/hunspell/" + name):
// exact classpath path only (missing → empty); parse with CachingWordListLoader line rules.
func loadHunspellWordList(t *testing.T, shortCode, name string) []string {
	t.Helper()
	path := spelling.DiscoverSpellingResource(shortCode + "/hunspell/" + name)
	if path == "" {
		return []string{}
	}
	f, err := os.Open(path)
	require.NoError(t, err, "open %s", path)
	defer f.Close()
	words, err := spelling.ParseWordListLines(f)
	require.NoError(t, err, "parse %s", path)
	return words
}

// wordListIntersection mirrors Java distinct(prohibit).filter(spelling.contains).
func wordListIntersection(prohibited, spellingWords []string) []string {
	spellingSet := make(map[string]struct{}, len(spellingWords))
	for _, w := range spellingWords {
		spellingSet[w] = struct{}{}
	}
	seen := make(map[string]struct{})
	var out []string
	for _, w := range prohibited {
		if _, dup := seen[w]; dup {
			continue
		}
		seen[w] = struct{}{}
		if _, ok := spellingSet[w]; ok {
			out = append(out, w)
		}
	}
	return out
}
