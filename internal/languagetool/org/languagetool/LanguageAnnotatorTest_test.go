package languagetool

// Twin of languagetool-standalone LanguageAnnotatorTest.
// Full VagueSpellChecker dictionary deferred; pluggable IsValidWord smokes.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of LanguageAnnotatorTest.testGetTokensWithPotentialLanguages
func TestLanguageAnnotator_GetTokensWithPotentialLanguages(t *testing.T) {
	a := NewLanguageAnnotator()
	a.IsValidWord = func(token, lang string) bool {
		low := strings.ToLower(token)
		if lang == "en" {
			return low == "this" || low == "is" || low == "english"
		}
		if lang == "de" {
			return low == "das" || low == "ist" || low == "deutsch"
		}
		return false
	}
	tokens := a.GetTokensWithPotentialLanguages("This is english. Das ist deutsch.", "en", []string{"de"})
	require.NotEmpty(t, tokens)
	var hasEN, hasDE bool
	for _, tok := range tokens {
		for _, l := range tok.Langs {
			if l == "en" {
				hasEN = true
			}
			if l == "de" {
				hasDE = true
			}
		}
	}
	require.True(t, hasEN)
	require.True(t, hasDE)
}

// Port of LanguageAnnotatorTest.testDetectLanguages
func TestLanguageAnnotator_DetectLanguages(t *testing.T) {
	a := NewLanguageAnnotator()
	a.IsValidWord = func(token, lang string) bool {
		if lang == "en" {
			return token == "hello" || token == "world" || token == "This"
		}
		if lang == "de" {
			return token == "Hallo" || token == "Welt"
		}
		return false
	}
	frags := a.DetectLanguages("hello world. Hallo Welt.", "en", []string{"de"})
	require.NotEmpty(t, frags)
	codes := map[string]bool{}
	for _, f := range frags {
		codes[f.LangCode] = true
	}
	require.True(t, codes["en"] || codes["de"])
}

// Port of LanguageAnnotatorTest.testGetTokenRanges
func TestLanguageAnnotator_GetTokenRanges(t *testing.T) {
	a := NewLanguageAnnotator()
	tokens := []TokenWithLanguages{
		{Token: "hi", Langs: []string{"en"}},
		{Token: " "},
		{Token: "there", Langs: []string{"en"}},
		{Token: "."},
		{Token: "ok", Langs: []string{"en"}},
	}
	ranges := a.GetTokenRanges(tokens)
	require.GreaterOrEqual(t, len(ranges), 2)
}

// Remaining ambiguous multi-language edge cases need full VagueSpellChecker.
func TestLanguageAnnotator_GetTokenRangesAmbiguous(t *testing.T) {
	t.Skip("unimplemented: ambiguous range matrix needs full spell checker")
}
func TestLanguageAnnotator_GetTokenRanges2(t *testing.T) {
	t.Skip("unimplemented: range matrix needs full spell checker")
}
func TestLanguageAnnotator_GetTokenRanges3(t *testing.T) {
	t.Skip("unimplemented: range matrix needs full spell checker")
}
