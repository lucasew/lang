package en

// Twin of org.languagetool.rules.en.LanguageSpecificSpellcheckerTest
//   → SpellcheckerTest.runLanguageSpecificTest()
//
// Java (SpellcheckerTest.runLanguageSpecificTest):
//   if Languages.get().isEmpty() → RuntimeException
//   for each Language lang in Languages.get():
//     prohibitedWords = CachingWordListLoader.loadWords(shortCode + "/hunspell/prohibit.txt")
//     spellingWords   = CachingWordListLoader.loadWords(shortCode + "/hunspell/spelling.txt")
//     intersection(prohibited distinct ∩ spelling) must be empty
//   if Languages.get().size() > 5 && totalProhibited == 0 → RuntimeException
//
// EN module test only invokes runLanguageSpecificTest(). Official resource paths match
// Java DataBroker (en/hunspell/{prohibit,spelling}.txt). Missing lists → empty (Java
// resourceExists false), no invent. EN lists are asserted present when the submodule is
// checked out (honest skip if not).

import (
	"os"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/stretchr/testify/require"
)

// Port of LanguageSpecificSpellcheckerTest.testRules → runLanguageSpecificTest().
func TestLanguageSpecificSpellchecker_Rules(t *testing.T) {
	runLanguageSpecificTest(t)
}

// runLanguageSpecificTest ports SpellcheckerTest.runLanguageSpecificTest.
func runLanguageSpecificTest(t *testing.T) {
	t.Helper()

	// EN module classpath always has English resources; skip honestly if official files
	// are absent (submodule not checked out) — no invent lists.
	// Exact Java classpath paths only (no DiscoverLangHunspellWordList fallbacks to
	// spelling/ or lang-root layouts — CachingWordListLoader uses resourceExists on
	// shortCode+"/hunspell/"+name only).
	enProhibitPath := spelling.DiscoverSpellingResource("en/hunspell/prohibit.txt")
	enSpellingPath := spelling.DiscoverSpellingResource("en/hunspell/spelling.txt")
	if enProhibitPath == "" || enSpellingPath == "" {
		t.Skip("official en/hunspell/prohibit.txt or spelling.txt not discoverable — honest skip, no invent")
	}

	languagetool.EnsureBuiltInLanguagesRegistered()
	langs := languagetool.GlobalLanguages.Get()
	if len(langs) == 0 {
		// Java: throw new RuntimeException("Warning: no languages found in classpath - cannot run any spell check tests");
		t.Fatal("Warning: no languages found in classpath - cannot run any spell check tests")
	}

	totalProhibited := 0
	for _, lang := range langs {
		short := lang.GetShortCode()
		// Java: new CachingWordListLoader().loadWords(short + "/hunspell/{prohibit,spelling}.txt")
		// Missing resource → empty list (CachingWordListLoader resourceExists false).
		prohibitedWords := loadHunspellWordList(t, short, "prohibit.txt")
		spellingWords := loadHunspellWordList(t, short, "spelling.txt")
		totalProhibited += len(prohibitedWords)

		t.Logf("Testing %s: %d words from prohibit.txt and %d words from spelling.txt",
			lang.Name, len(prohibitedWords), len(spellingWords))

		// Java: prohibitedWords.stream().distinct().filter(spellingWords::contains).collect(toSet())
		spellingSet := make(map[string]struct{}, len(spellingWords))
		for _, w := range spellingWords {
			spellingSet[w] = struct{}{}
		}
		seen := make(map[string]struct{})
		var intersection []string
		for _, w := range prohibitedWords {
			if _, dup := seen[w]; dup {
				continue
			}
			seen[w] = struct{}{}
			if _, ok := spellingSet[w]; ok {
				intersection = append(intersection, w)
			}
		}
		if len(intersection) > 0 {
			// Java: throw new RuntimeException("Word(s) appear in both spelling.txt and prohibit.txt - this doesn't make sense: " + intersection);
			t.Fatalf("Word(s) appear in both spelling.txt and prohibit.txt - this doesn't make sense: %v", intersection)
		}
	}

	// Sector outcomes for English short code: lists load and prohibit is non-empty.
	enProhibited := loadHunspellWordList(t, "en", "prohibit.txt")
	enSpelling := loadHunspellWordList(t, "en", "spelling.txt")
	require.NotEmpty(t, enProhibited, "en/hunspell/prohibit.txt must load non-empty official words")
	require.GreaterOrEqual(t, len(enProhibited), 0)
	require.NotNil(t, enSpelling)
	// Intersection for en already covered in the Languages loop; re-assert for the short code.
	require.Empty(t, wordListIntersection(enProhibited, enSpelling),
		"en prohibit.txt ∩ spelling.txt must be empty")

	// Java: if (Languages.get().size() > 5 && totalProhibited == 0) throw ...
	if len(langs) > 5 && totalProhibited == 0 {
		t.Fatalf("No prohibited words at all for %d languages - is there a bug in this test?!", len(langs))
	}
}

// loadHunspellWordList ports CachingWordListLoader.loadWords(short + "/hunspell/" + name):
// exact classpath path only (missing → empty); parse with CachingWordListLoader line rules.
func loadHunspellWordList(t *testing.T, shortCode, name string) []string {
	t.Helper()
	// Java path is always shortCode+"/hunspell/"+name — no NL spelling/ or PT root fallback.
	path := spelling.DiscoverSpellingResource(shortCode + "/hunspell/" + name)
	if path == "" {
		return []string{}
	}
	f, err := os.Open(path)
	require.NoError(t, err, "open %s", path)
	defer f.Close()
	// ParseWordListLines = CachingWordListLoader.CacheLoader.load line loop (comments, Java trim).
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
