package hunspell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of HunspellRule.computeNonWordPattern — WORDCHARS from .aff.
// (Go: NonWordSplitter equivalent of Java Pattern with WORDCHARS lookahead.)
func TestComputeNonWordPattern_WORDCHARS(t *testing.T) {
	// da_DK.aff: WORDCHARS -.
	spl := ComputeNonWordSplitterFromString("SET UTF-8\nWORDCHARS -.\n")
	// Split should keep hyphen/period attached (not split on them as non-word).
	require.Equal(t, []string{"well-known", "word"}, spl.Split("well-known word"))
	require.Equal(t, []string{"file.txt"}, spl.Split("file.txt"))

	// Default (no WORDCHARS): hyphen splits
	def := ComputeNonWordSplitterFromString("SET UTF-8\n")
	require.Equal(t, []string{"well", "known"}, def.Split("well-known"))
}

func TestComputeNonWordPattern_RealGermanAff(t *testing.T) {
	wd, _ := os.Getwd()
	var aff string
	dir := wd
	for {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/hunspell/de_DE.aff")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			aff = cand
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if aff == "" {
		t.Skip("de_DE.aff not in tree")
	}
	f, err := os.Open(aff)
	require.NoError(t, err)
	defer f.Close()
	spl := ComputeNonWordSplitter(f)
	// de WORDCHARS ß-. — hyphen stays as word char for split
	require.Equal(t, []string{"Dampf-Schiff", "x"}, spl.Split("Dampf-Schiff x"))
}

// Twin of getDictFilenameInResources.
func TestGetDictFilenameInResources(t *testing.T) {
	require.Equal(t, "/de/hunspell/de_DE.dic", GetDictFilenameInResources("de", "de_DE"))
	require.Equal(t, "/da/hunspell/da_DK.dic", GetDictFilenameInResources("da", "da_DK"))
	require.Equal(t, "/de/hunspell/de_DE.dic", GetDictFilenameInResourcesFromLangCode("de-DE"))
	require.Equal(t, "/de/hunspell/de_AT.dic", GetDictFilenameInResourcesFromLangCode("de_AT"))
	require.Equal(t, "/da/hunspell/da.dic", GetDictFilenameInResourcesFromLangCode("da"))
}

// TokenizeText uses NonWordSplitter when set.
func TestTokenizeText_WORDCHARS(t *testing.T) {
	r := NewHunspellRule("da", NewMapHunspellDictionary(nil))
	r.NonWordSplitter = ComputeNonWordSplitterFromString("WORDCHARS -.\n")
	require.Equal(t, []string{"well-known", "word"}, r.TokenizeText("well-known word"))
}

// isQuotedCompound base is false.
func TestIsQuotedCompound_DefaultFalse(t *testing.T) {
	r := NewHunspellRule("en", nil)
	require.False(t, r.IsQuotedCompound(nil, 1, "\"foo"))
	r.IsQuotedCompoundFn = func(s *languagetool.AnalyzedSentence, idx int, token string) bool {
		return strings.HasPrefix(token, "\"")
	}
	require.True(t, r.IsQuotedCompound(nil, 1, "\"foo"))
}

// GetSentenceTextWithoutUrlsAndImmunizedTokens: immunized → spaces of UTF-16 len.
func TestGetSentenceTextWithoutUrlsAndImmunizedTokens(t *testing.T) {
	r := NewHunspellRule("en", NewMapHunspellDictionary([]string{"hello", "world"}))
	sent := languagetool.AnalyzePlain("hello world")
	text := r.GetSentenceTextWithoutUrlsAndImmunizedTokens(sent)
	require.Contains(t, text, "hello")
	require.Contains(t, text, "world")

	sent2 := languagetool.AnalyzePlain("hello world")
	for _, tok := range sent2.GetTokens() {
		if tok != nil && tok.GetToken() == "hello" {
			tok.Immunize(0)
		}
	}
	text2 := r.GetSentenceTextWithoutUrlsAndImmunizedTokens(sent2)
	require.NotContains(t, strings.TrimSpace(text2), "hello")
	require.Contains(t, text2, "world")
	require.Contains(t, text2, "     ") // 5 spaces for "hello"
}

// URL replaced with spaces.
func TestGetSentenceText_URLBlanked(t *testing.T) {
	r := NewHunspellRule("en", NewMapHunspellDictionary([]string{"see"}))
	sent := languagetool.AnalyzePlain("see http://example.com")
	text := r.GetSentenceTextWithoutUrlsAndImmunizedTokens(sent)
	require.Contains(t, text, "see")
	require.NotContains(t, text, "http")
}

// NoSuggestion: SuggestFn empty; Match has no suggestions.
func TestHunspellNoSuggestion_SuggestFn(t *testing.T) {
	dict := NewMapHunspellDictionary([]string{"hello"})
	dict.SetSuggestions("helo", []string{"hello"})
	r := NewHunspellNoSuggestionRule("en", dict)
	require.Empty(t, r.Suggest("helo"))
	require.Empty(t, r.HunspellRule.Suggest("helo")) // via SuggestFn
	ms, err := r.Match(languagetool.AnalyzePlain("helo"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
	require.Empty(t, ms[0].GetSuggestedReplacements())
}
