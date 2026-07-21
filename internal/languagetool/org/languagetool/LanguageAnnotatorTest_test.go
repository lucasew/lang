package languagetool

// Twin of languagetool-standalone LanguageAnnotatorTest.
import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func enDeValid(token, lang string) bool {
	low := strings.ToLower(token)
	en := map[string]bool{
		"this": true, "is": true, "a": true, "test": true, "an": true, "english": true,
		"new": true, "bicycle": true, "text": true, "one": true, "two": true, "three": true,
		"use": true, "it": true, "on": true, "our": true, "website": true, "nun": true,
	}
	de := map[string]bool{
		"der": true, "große": true, "haus": true, "hier": true, "kommt": true, "ein": true,
		"deutscher": true, "satz": true, "nun": true, "steht": true, "was": true, "auf": true,
		"deutsch": true, "geht": true, "es": true, "weiter": true, "nutzen": true, "sie": true,
		"unserer": true, "webseite": true, "a": true,
	}
	if lang == "en" || lang == "en-US" {
		return en[low]
	}
	if lang == "de" || lang == "de-DE" {
		return de[low]
	}
	return false
}

func TestLanguageAnnotator_GetTokensWithPotentialLanguages(t *testing.T) {
	a := NewLanguageAnnotator()
	a.IsValidWord = enDeValid
	tokens := a.GetTokensWithPotentialLanguages("This is a new bicycle.", "en", []string{"de"})
	require.NotEmpty(t, tokens)
	found := false
	for _, tok := range tokens {
		if tok.Token == "This" {
			require.Contains(t, tok.Langs, "en")
			found = true
		}
	}
	require.True(t, found)
}

func TestLanguageAnnotator_DetectLanguages(t *testing.T) {
	a := NewLanguageAnnotator()
	a.IsValidWord = enDeValid
	frags := a.DetectLanguages("This is an English test. Hier kommt ein deutscher Satz.", "en", []string{"de"})
	require.NotEmpty(t, frags)
}

func TestLanguageAnnotator_GetTokenRanges(t *testing.T) {
	a := NewLanguageAnnotator()
	tokens := []TokenWithLanguages{
		{Token: "This", Langs: []string{"en"}},
		{Token: " "},
		{Token: "is", Langs: []string{"en"}},
		{Token: " "},
		{Token: "a", Langs: []string{"en"}},
		{Token: " "},
		{Token: "test", Langs: []string{"en"}},
		{Token: "!"},
		{Token: "Hier", Langs: []string{"de"}},
		{Token: " "},
		{Token: "geht", Langs: []string{"de"}},
		{Token: " "},
		{Token: "es", Langs: []string{"de"}},
		{Token: " "},
		{Token: "weiter", Langs: []string{"de"}},
		{Token: "."},
	}
	ranges := a.GetTokenRanges(tokens)
	require.GreaterOrEqual(t, len(ranges), 2)
	s := TokenRangeString(ranges)
	require.Contains(t, s, "This")
	require.Contains(t, s, "Hier")
	labeled := a.GetTokenRangesWithLang(ranges, "en", []string{"de"})
	require.Len(t, labeled, len(ranges))
	require.NotEmpty(t, labeled[0].Lang)
}

func TestLanguageAnnotator_GetTokenRangesAmbiguous(t *testing.T) {
	a := NewLanguageAnnotator()
	a.IsValidWord = enDeValid
	tokens := a.GetTokensWithPotentialLanguages("a. Hier geht es weiter.", "de", []string{"en"})
	ranges := a.GetTokenRanges(tokens)
	labeled := a.GetTokenRangesWithLang(ranges, "de", []string{"en"})
	require.NotEmpty(t, labeled)
}

func TestLanguageAnnotator_GetTokenRanges2(t *testing.T) {
	a := NewLanguageAnnotator()
	tokens := []TokenWithLanguages{
		{Token: "This", Langs: []string{"en"}},
		{Token: " "},
		{Token: "is", Langs: []string{"en"}},
		{Token: " "},
		{Token: "a", Langs: []string{"en"}},
		{Token: " "},
		{Token: "test", Langs: []string{"en"}},
		{Token: "."},
		{Token: "\""},
		{Token: "Hier", Langs: []string{"de"}},
		{Token: " "},
		{Token: "geht", Langs: []string{"de"}},
		{Token: " "},
		{Token: "es", Langs: []string{"de"}},
		{Token: " "},
		{Token: "weiter", Langs: []string{"de"}},
		{Token: "\""},
	}
	ranges := a.GetTokenRanges(tokens)
	require.GreaterOrEqual(t, len(ranges), 2)
}

func TestLanguageAnnotator_GetTokenRanges3(t *testing.T) {
	a := NewLanguageAnnotator()
	tokens := []TokenWithLanguages{
		{Token: "This", Langs: []string{"en"}},
		{Token: " "},
		{Token: "is", Langs: []string{"en"}},
		{Token: " "},
		{Token: "a", Langs: []string{"en"}},
		{Token: " "},
		{Token: "test", Langs: []string{"en"}},
		{Token: "."},
		{Token: "\""},
		{Token: "Hier", Langs: []string{"de"}},
		{Token: " "},
		{Token: "geht", Langs: []string{"de"}},
		{Token: " "},
		{Token: "es", Langs: []string{"de"}},
		{Token: " "},
		{Token: "weiter", Langs: []string{"de"}},
		{Token: "."},
		{Token: "\""},
	}
	ranges := a.GetTokenRanges(tokens)
	require.GreaterOrEqual(t, len(ranges), 2)
}

func TestLanguageAnnotator_DefaultWordTokenizer(t *testing.T) {
	// Java: mainLang.getWordTokenizer().tokenize — WordTokenizer keeps space tokens.
	a := NewLanguageAnnotator()
	toks := a.tokenize("Hello, world.")
	require.Equal(t, []string{"Hello", ",", " ", "world", "."}, toks)
}
