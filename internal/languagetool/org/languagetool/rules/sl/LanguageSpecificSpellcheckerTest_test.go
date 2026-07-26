package sl

// Twin of org.languagetool.rules.sl.LanguageSpecificSpellcheckerTest
//   → SpellcheckerTest.runLanguageSpecificTest()
//
// Java (SpellcheckerTest.runLanguageSpecificTest) for each Language on classpath:
//   prohibitedWords = CachingWordListLoader.loadWords(shortCode + "/hunspell/prohibit.txt")
//   spellingWords   = CachingWordListLoader.loadWords(shortCode + "/hunspell/spelling.txt")
//   intersection(prohibited distinct ∩ spelling) must be empty
//
// SL module test only invokes runLanguageSpecificTest(). SL-visible paths:
//   sl/hunspell/spelling.txt (present; comment-only upstream → empty word list)
//   sl/hunspell/prohibit.txt (absent → empty list, Java resourceExists false)
// Missing resources → empty (no invent). Intersection for SL must be empty.

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

// Port of LanguageSpecificSpellcheckerTest.testRules → runLanguageSpecificTest (SL sector).
func TestLanguageSpecificSpellchecker_Rules(t *testing.T) {
	// Official SL spelling list path (Java DataBroker sl/hunspell/spelling.txt).
	slSpellingPath := spelling.DiscoverSpellingResource("sl/hunspell/spelling.txt")
	if slSpellingPath == "" {
		t.Skip("official sl/hunspell/spelling.txt not discoverable — honest skip, no invent")
	}

	// SL-visible outcomes of SpellcheckerTest for short code "sl".
	// Java CachingWordListLoader: missing prohibit.txt → empty list (not invent).
	prohibitedWords := loadHunspellWordList(t, "sl", "prohibit.txt")
	spellingWords := loadHunspellWordList(t, "sl", "spelling.txt")
	t.Logf("Testing sl: %d words from prohibit.txt and %d words from spelling.txt",
		len(prohibitedWords), len(spellingWords))
	require.Empty(t, wordListIntersection(prohibitedWords, spellingWords),
		"sl prohibit.txt ∩ spelling.txt must be empty")

	// Speller rule surface still loads (LanguageSpecificSpellcheckerTest lives next to the rule).
	r := NewMorfologikSlovenianSpellerRule()
	require.Equal(t, MorfologikSlovenianSpellerRuleID, r.GetID())
	require.Equal(t, SlovenianSpellerDict, r.GetFileName())
	if morfologik.DiscoverLanguageDict(SlovenianSpellerDict) != "" {
		require.NotNil(t, r.Speller1)
	}

	// Analyze surface (JLanguageTool sl).
	require.NotEmpty(t, languagetool.NewJLanguageTool("sl").Analyze("test"))
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
