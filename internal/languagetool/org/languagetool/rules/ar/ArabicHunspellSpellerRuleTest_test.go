package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicHunspellSpellerRuleTest.java
//
// Full testRuleWithArabic match matrix requires Java DumontsHunspellDictionary
// (native hunspell.spell with ar.dic + ar.aff). Pure-Go FileHunspellDictionary loads
// .dic stems only (affix rules ignored) — same dependency leaf as blocked
// 3.C.2-VagueSpellChecker / 3.C.3-GermanSpellerRule.
// MapHunspell inject + SetSuggestions is invent and is not used as ready proof.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/hunspell"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Unit: Java ArabicHunspellSpellerRule overrides (id, path, latin, tokenize, tashkeel).
func TestArabicHunspellSpellerRule_Overrides(t *testing.T) {
	// Small map only exercises AR method control flow (tashkeel strip), not full Java twin.
	dict := hunspell.NewMapHunspellDictionary([]string{"كتب", "كتاب", "مقدمة"})
	r := NewArabicHunspellSpellerRule(dict)

	require.Equal(t, ArabicHunspellRuleID, r.GetID())
	require.Equal(t, "HUNSPELL_RULE_AR", r.GetID())
	require.Equal(t, ArabicHunspellDictPath, r.GetDictFilenameInResources("ar"))
	require.Equal(t, "/ar/hunspell/ar.dic", r.GetDictFilenameInResources("ar-SA"))
	require.False(t, r.IsLatinScript())
	require.True(t, r.NonLatinScript)

	// isMisspelled: strip tashkeel then super
	require.False(t, r.IsMisspelledStripped("كتب"))
	require.False(t, r.IsMisspelledStripped("كَتَبَ")) // → كتب
	require.False(t, r.IsMisspelledStripped("مُقَدِّمَةُ")) // → مقدمة
	require.True(t, r.IsMisspelledStripped("xyzzy"))
	require.True(t, r.IsMisspelledStripped("عليييكم"))

	// ignoreWord: strip tashkeel then super (ignore list / no-letter gates)
	r.AddIgnoreWords("كتب")
	require.True(t, r.IgnoreWord("كتب"))
	require.True(t, r.IgnoreWord("كَتَبَ"), "tashkeel form ignored after strip")
	require.False(t, r.IgnoreWord("كتابx"))

	// tokenizeText: tashkeel kept with letters; punctuation splits
	toks := r.TokenizeText("مرحبا، عالم!")
	require.Equal(t, []string{"مرحبا", "عالم"}, filterEmpty(toks))
	// diacritics stay inside token (not split points)
	toks = TokenizeArabicSpellText("كَتَبَ.")
	require.Contains(t, toks, "كَتَبَ")
	require.Equal(t, tools.RemoveTashkeel("كَتَبَ"), "كتب")

	// Match path uses stripped isMisspelled
	m, err := r.Match(languagetool.AnalyzePlain("كَتَبَ"))
	require.NoError(t, err)
	require.Empty(t, m)
}

// Official classpath load of ar.dic (+ companion ar.aff path; affix not expanded).
func TestArabicHunspellSpellerRule_OfficialDicLoads(t *testing.T) {
	dict := hunspell.TryOpenFromClasspath(ArabicHunspellDictPath)
	if dict == nil {
		t.Skip("official ar.dic not discoverable")
	}
	r := NewArabicHunspellSpellerRule(dict)
	require.Equal(t, ArabicHunspellRuleID, r.GetID())
	require.Equal(t, ArabicHunspellDictPath, r.GetDictFilenameInResources(""))

	// Stems present as raw .dic entries (no affix needed).
	require.False(t, r.IsMisspelledStripped("عليكم"))
	require.False(t, r.IsMisspelledStripped("تصفح"))
	require.False(t, r.IsMisspelledStripped("هذه"))
	require.False(t, r.IsMisspelledStripped("إذا"))
	require.False(t, r.IsMisspelledStripped("مقدمة"))
	// tashkeel strip hits stem
	require.False(t, r.IsMisspelledStripped("مُقَدِّمَةُ"))

	// URL tokens ignored (IsUrl / CanBeIgnoredToken — not invent soft-skip)
	m, err := r.Match(languagetool.AnalyzePlain("تصفح http://foo.org/bar."))
	require.NoError(t, err)
	require.Empty(t, m, "Java ignores URL; تصفح is a dic stem")
}

// Documents Java testRuleWithArabic cases that pure-Go cannot twin without affix engine.
// Do not invent MapHunspell goldens to greenwash this matrix.
func TestArabicHunspellSpellerRule_RuleWithArabic_AffixBlocked(t *testing.T) {
	dict := hunspell.TryOpenFromClasspath(ArabicHunspellDictPath)
	if dict == nil {
		t.Skip("official ar.dic not discoverable")
	}
	r := NewArabicHunspellSpellerRule(dict)

	// Java assertEquals(0, match("السلام عليكم.")) — needs affix form السلام from سلام+ال.
	// Pure-Go stem map flags السلام as misspelled.
	require.True(t, r.IsMisspelledStripped("السلام"),
		"السلام requires .aff expansion; FileHunspellDictionary stems-only")
	require.True(t, r.IsMisspelledStripped("والبلاد"),
		"والبلاد requires .aff expansion")
	require.True(t, r.IsMisspelledStripped("العبارة"),
		"العبارة requires .aff expansion (Java suggestion target for العباره)")

	// Misspell surfaces still miss (stem-or-not); suggestions need native hunspell.suggest.
	m, err := r.Match(languagetool.AnalyzePlain("تظاعف"))
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Empty(t, m[0].GetSuggestedReplacements(),
		"no invent SetSuggestions; pure-Go FileHunspell Suggest is empty without native engine")

	m, err = r.Match(languagetool.AnalyzePlain("هذه العباره فيها أغلاط."))
	require.NoError(t, err)
	// Java: exactly 1 match (العباره); pure-Go may also false-flag affix-only tokens if present.
	// With stems هذه/فيها/أغلاط present, only العباره should miss among those four —
	// but we do not claim full position/suggestion parity here.
	require.NotEmpty(t, m)
	found := false
	for _, rm := range m {
		if rm.GetFromPos() == 4 && rm.GetToPos() == 11 {
			found = true
			require.Empty(t, rm.GetSuggestedReplacements(), "no invent العبارة suggestion")
		}
	}
	require.True(t, found, "العباره span 4..11 should be flagged (surface misspelling)")

	// Correct multi-token sentence Java expects 0 matches — pure-Go false-positives affix forms.
	m, err = r.Match(languagetool.AnalyzePlain("السلام عليكم."))
	require.NoError(t, err)
	require.NotEmpty(t, m, "blocked: Java 0 matches; pure-Go flags affix form السلام")
}

// filterEmpty drops intermediate empties from Pattern.split-style results for assertion convenience.
func filterEmpty(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
