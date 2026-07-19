package spelling

import (
	"testing"

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
	require.Contains(t, words, "kr")
}

func TestApplyDefaultSpellingWordLists_Ignore(t *testing.T) {
	if DiscoverLangHunspellWordList("da", "ignore.txt") == "" {
		t.Skip("da ignore.txt not in tree")
	}
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "da")
	r.IsMisspelled = func(string) bool { return true }
	ApplyDefaultSpellingWordLists(r)
	require.True(t, r.AcceptWord("kr"))
	require.False(t, r.AcceptWord("xyzzyqqqnotaword"))
}

func TestDiscoverSpellingGlobal(t *testing.T) {
	p := DiscoverSpellingGlobal()
	if p == "" {
		t.Skip("spelling_global.txt not in tree")
	}
	require.Contains(t, p, "spelling_global.txt")
	words, err := LoadSpellingWordListFile(p)
	require.NoError(t, err)
	require.NotEmpty(t, words)
	// Official global single-token entries (see spelling_global.txt)
	require.Contains(t, words, "log4j")
	require.Contains(t, words, "mp3tag")
}

func TestApplyDefaultSpellingWordLists_GlobalSpelling(t *testing.T) {
	if DiscoverSpellingGlobal() == "" {
		t.Skip("spelling_global.txt not in tree")
	}
	// Language-local lists optional; global must still apply.
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "da")
	r.IsMisspelled = func(string) bool { return true }
	ApplyDefaultSpellingWordLists(r)
	// Single-token global words accepted (Java wordsToBeIgnored from GLOBAL_SPELLING_FILE).
	require.True(t, r.AcceptWord("log4j"))
	require.True(t, r.AcceptWord("mp3tag"))
	require.False(t, r.AcceptWord("xyzzyqqqnotaword"))
}

func TestApplyDefaultSpellingWordLists_GlobalEvenWithoutLangCode(t *testing.T) {
	if DiscoverSpellingGlobal() == "" {
		t.Skip("spelling_global.txt not in tree")
	}
	// LanguageCode empty: still load global (all-language file).
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "")
	r.IsMisspelled = func(string) bool { return true }
	ApplyDefaultSpellingWordLists(r)
	require.True(t, r.AcceptWord("log4j"))
}

func TestLanguageVariantSpellingClasspath(t *testing.T) {
	require.Equal(t, "en/hunspell/spelling_en-US.txt", LanguageVariantSpellingClasspath("en"))
	require.Equal(t, "en/hunspell/spelling_en-US.txt", LanguageVariantSpellingClasspath("en-US"))
	require.Equal(t, "en/hunspell/spelling_en-GB.txt", LanguageVariantSpellingClasspath("en-GB"))
	require.Equal(t, "en/hunspell/spelling_en-CA.txt", LanguageVariantSpellingClasspath("en-CA"))
	require.Equal(t, "de/hunspell/spelling-de-AT.txt", LanguageVariantSpellingClasspath("de-AT"))
	require.Equal(t, "de/hunspell/spelling-de-CH.txt", LanguageVariantSpellingClasspath("de-CH"))
	require.Empty(t, LanguageVariantSpellingClasspath("de"))
	require.Empty(t, LanguageVariantSpellingClasspath("pl"))
}

func TestApplyVariantSpellingFile_American(t *testing.T) {
	rel := "en/hunspell/spelling_en-US.txt"
	if DiscoverSpellingResource(rel) == "" {
		t.Skip("spelling_en-US.txt not in tree")
	}
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en-US")
	r.IsMisspelled = func(string) bool { return true }
	ApplyVariantSpellingFile(r, rel)
	// official en-US variant list (see spelling_en-US.txt)
	require.True(t, r.AcceptWord("disulfide") || r.AcceptWord("micrometer") || r.AcceptWord("lockup"),
		"expected a word from spelling_en-US.txt in ignore set, size=%d", len(r.IgnoreWords))
	require.False(t, r.AcceptWord("xyzzyqqqnotaword"))
}

func TestApplyDefaultSpellingWordLists_IncludesVariant(t *testing.T) {
	if DiscoverSpellingResource("en/hunspell/spelling_en-GB.txt") == "" {
		t.Skip("spelling_en-GB.txt not in tree")
	}
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_GB", "spell", "en-GB")
	r.IsMisspelled = func(string) bool { return true }
	ApplyDefaultSpellingWordLists(r)
	// variant file loaded — set non-empty when file has entries
	require.NotEmpty(t, r.IgnoreWords)
}

func TestDiscoverLangHunspellWordList_DutchSpellingDirFallback(t *testing.T) {
	// Java MorfologikDutchSpellerRule: /nl/spelling/ignore.txt (not hunspell/)
	p := DiscoverLangHunspellWordList("nl", "ignore.txt")
	if p == "" {
		t.Skip("nl spelling ignore not in tree")
	}
	require.Contains(t, p, "spelling")
	require.Contains(t, p, "ignore.txt")
	words, err := LoadSpellingWordListFile(p)
	require.NoError(t, err)
	require.Contains(t, words, "'s-avonds")
}

func TestApplyDefaultSpellingWordLists_DutchSpellingDir(t *testing.T) {
	if DiscoverLangHunspellWordList("nl", "ignore.txt") == "" {
		t.Skip("nl spelling ignore not in tree")
	}
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_NL_NL", "spell", "nl")
	r.IsMisspelled = func(string) bool { return true }
	ApplyDefaultSpellingWordLists(r)
	// "Abra" is a single-token line in official nl/spelling/ignore.txt
	require.True(t, r.IgnoreWord("Abra"))
	// hyphenated lines become MultiWordIgnore (DutchWordTokenizer splits on '-')
	require.NotEmpty(t, r.MultiWordIgnore)
}

func TestDiscoverLangHunspellWordList_PortugueseRootFallback(t *testing.T) {
	// Java MorfologikPortugueseSpellerRule: "pt/ignore.txt" (resource root)
	p := DiscoverLangHunspellWordList("pt", "ignore.txt")
	if p == "" {
		t.Skip("pt/ignore.txt not in tree")
	}
	require.Contains(t, p, "ignore.txt")
	require.NotContains(t, p, "hunspell")
	words, err := LoadSpellingWordListFile(p)
	require.NoError(t, err)
	require.Contains(t, words, "ignorewordoogaboogatest")
}

func TestApplyDefaultSpellingWordLists_PortugueseRoot(t *testing.T) {
	if DiscoverLangHunspellWordList("pt", "ignore.txt") == "" {
		t.Skip("pt/ignore.txt not in tree")
	}
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_PT_PT", "spell", "pt")
	r.IsMisspelled = func(string) bool { return true }
	ApplyDefaultSpellingWordLists(r)
	require.True(t, r.IgnoreWord("ignorewordoogaboogatest"))
	if DiscoverLangHunspellWordList("pt", "prohibit.txt") != "" {
		require.True(t, r.IsProhibited("prohibitwordoogaboogatest"))
	}
}

func TestApplyDefaultSpellingWordLists_CatalanAdditional(t *testing.T) {
	// MorfologikCatalanSpellerRule.getAdditionalSpellingFileNames: multiwords + spelling-special
	if DiscoverSpellingResource("ca/spelling-special.txt") == "" && DiscoverSpellingResource("ca/multiwords.txt") == "" {
		t.Skip("ca multiwords/spelling-special not in tree")
	}
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_CA_ES", "spell", "ca")
	r.IsMisspelled = func(string) bool { return true }
	ApplyDefaultSpellingWordLists(r)
	// spelling-special contains inalàmbric (single-token ignore)
	if DiscoverSpellingResource("ca/spelling-special.txt") != "" {
		require.True(t, r.IgnoreWord("inalàmbric"), "ca/spelling-special.txt should load as ignore")
	}
}

func TestApplySpellingResourcePaths_SerbianEkavian(t *testing.T) {
	// Java MorfologikEkavianSpellerRule: ignored.txt under dictionary/ekavian (not ignore.txt)
	if DiscoverSpellingResource("sr/dictionary/ekavian/ignored.txt") == "" {
		t.Skip("sr/dictionary/ekavian/ignored.txt not in tree")
	}
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_SR_EKAVIAN", "spell", "sr")
	r.IsMisspelled = func(string) bool { return true }
	ApplySpellingResourcePaths(r,
		"sr/dictionary/ekavian/ignored.txt",
		"sr/dictionary/ekavian/spelling.txt",
		"sr/dictionary/ekavian/prohibit.txt", // may be missing — no-op
	)
	// files may be comment-only; ensure call does not invent and discover works
	require.NotEmpty(t, DiscoverSpellingResource("sr/dictionary/ekavian/spelling.txt"))
}

func TestDiscoverSpellingResource_SerbianPaths(t *testing.T) {
	p := DiscoverSpellingResource("sr/dictionary/jekavian/ignored.txt")
	if p == "" {
		t.Skip("sr jekavian ignored not in tree")
	}
	require.Contains(t, p, "jekavian")
	require.Contains(t, p, "ignored.txt")
}
