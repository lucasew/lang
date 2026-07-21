package spelling

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestIgnoreWord_NoLetterAndMaxLength(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	require.True(t, r.IgnoreWord("123"))
	require.True(t, r.IgnoreWord("..."))
	// oversize
	long := make([]byte, MaxTokenLength+1)
	for i := range long {
		long[i] = 'a'
	}
	require.True(t, r.IgnoreWord(string(long)))
}

func TestIgnoreWord_LatinScriptIgnoresNonLatin(t *testing.T) {
	// Java isLatinScript() default true: pHasNoLetterLatin → pure Cyrillic ignored
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en")
	require.True(t, r.IgnoreWord("привет"), "no Latin letters → ignore on Latin-script rule")
	require.False(t, r.IgnoreWord("hello"))
	require.False(t, r.IgnoreWord("café")) // Latin + combining still Latin script
}

func TestIgnoreWord_NonLatinScriptAnyLetter(t *testing.T) {
	// Java isLatinScript() false: only ignore when no \p{L} at all
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_RU_RU", "spell", "ru")
	r.NonLatinScript = true
	require.False(t, r.IgnoreWord("привет"), "Cyrillic has letters → not ignored")
	require.False(t, r.IgnoreWord("hello"), "Latin has letters → not ignored via letter gate")
	require.True(t, r.IgnoreWord("123"))
}

func TestIgnoreWord_TrailingPeriod(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AddIgnoreWords("LanguageTool")
	require.True(t, r.IgnoreWord("LanguageTool."))
	require.False(t, r.IgnoreWord("OtherTool."))
}

func TestIsIgnoredNoCase_ConvertsCase(t *testing.T) {
	// Java isIgnoredNoCase: convertsCase lower-match only when !isMixedCase.
	// isMixedCase = !allUpper && !capitalized(firstUpper+restLower) && !allLower
	// so "LanguageTool" is mixed; "Languagetool"/"LANGUAGETOOL" are not.
	r := NewSpellingCheckRule("MORFOLOGIK_RULE_EN_US", "spell", "en")
	r.AddIgnoreWords("languagetool")
	r.ConvertsCase = true
	require.True(t, r.IsIgnoredNoCase("languagetool"))
	require.True(t, r.IsIgnoredNoCase("Languagetool")) // capitalized, not mixed
	require.True(t, r.IsIgnoredNoCase("LANGUAGETOOL")) // all upper, not mixed
	// mixed case: Java does not lower-match (StringTools.isMixedCase)
	require.False(t, r.IsIgnoredNoCase("LanguageTool"))
	require.False(t, r.IsIgnoredNoCase("LanguageTOOL"))
	r.ConvertsCase = false
	require.False(t, r.IsIgnoredNoCase("Languagetool"))
	require.False(t, r.IsIgnoredNoCase("LANGUAGETOOL"))
}

func TestStartsWithIgnoredWord(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AddIgnoreWords("LanguageTool", "foo", "foobar")
	// length < 4 → 0
	require.Equal(t, 0, r.StartsWithIgnoredWord("foo", true))
	// exact long word
	require.Equal(t, javaStringLenSpell("LanguageTool"), r.StartsWithIgnoredWord("LanguageTool", true))
	// prefix: LanguageToolish starts with LanguageTool
	n := r.StartsWithIgnoredWord("LanguageToolish", true)
	require.Equal(t, javaStringLenSpell("LanguageTool"), n)
	// no match
	require.Equal(t, 0, r.StartsWithIgnoredWord("xyzzytool", true))
}

// Java word.length / commonPrefix are UTF-16; é vs è diverge after first unit of the accented char.
func TestStartsWithIgnoredWord_AccentedUTF16(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "fr")
	r.AddIgnoreWords("caféteria")
	// exact
	require.Equal(t, javaStringLenSpell("caféteria"), r.StartsWithIgnoredWord("caféteria", true))
	// longer word with ignored prefix
	require.Equal(t, javaStringLenSpell("caféteria"), r.StartsWithIgnoredWord("caféteriaxyz", true))
	// short gate: length 3 UTF-16
	require.Equal(t, 0, r.StartsWithIgnoredWord("caf", true))
}

func TestIgnoreWordsWithLength_UTF16(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.IgnoreWordsWithLength = 1
	// Java: word.length() <= 1 — single BMP letter
	require.True(t, r.IsIgnoredNoCase("a"))
	// emoji is 2 UTF-16 units → not ignored by length-1
	require.False(t, r.IsIgnoredNoCase("😀"))
	r.IgnoreWordsWithLength = 2
	require.True(t, r.IsIgnoredNoCase("😀"))
}

func TestIgnoreToken(t *testing.T) {
	r := NewSpellingCheckRule("HUNSPELL_RULE", "spell", "en")
	r.AddIgnoreWords("ok")
	sent := languagetool.AnalyzePlain("ok bad")
	tokens := sent.GetTokensWithoutWhitespace()
	// find indices
	for i, tok := range tokens {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		if tok.GetToken() == "ok" {
			require.True(t, r.IgnoreToken(tokens, i))
		}
		if tok.GetToken() == "bad" {
			require.False(t, r.IgnoreToken(tokens, i))
		}
	}
}

func TestIgnorePotentiallyMisspelledWord_DefaultFalse(t *testing.T) {
	r := NewSpellingCheckRule("MORFOLOGIK_RULE", "spell", "en")
	require.False(t, r.IgnorePotentiallyMisspelledWord("anything"))
	r.IgnorePotentiallyMisspelledWordFn = func(word string) bool {
		return word == "compoundok"
	}
	require.True(t, r.IgnorePotentiallyMisspelledWord("compoundok"))
	require.False(t, r.IgnorePotentiallyMisspelledWord("stillbad"))
}
