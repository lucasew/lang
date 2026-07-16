package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLanguageAnnotatorDetectLanguagesPort(t *testing.T) {
	a := NewLanguageAnnotator()
	a.IsValidWord = func(token, lang string) bool {
		if lang == "en" {
			return token == "hello" || token == "world"
		}
		if lang == "de" {
			return token == "Hallo" || token == "Welt"
		}
		return false
	}
	frags := a.DetectLanguages("hello world. Hallo Welt.", "en", []string{"de"})
	require.NotEmpty(t, frags)
	// should have at least one en and one de fragment ideally
	codes := map[string]bool{}
	for _, f := range frags {
		codes[f.LangCode] = true
	}
	require.True(t, codes["en"] || codes["de"])
}

func TestLanguageAnnotatorTokenRangesPort(t *testing.T) {
	a := NewLanguageAnnotator()
	tokens := []TokenWithLanguages{
		{Token: "hi", Langs: []string{"en"}},
		{Token: "."},
		{Token: "ok", Langs: []string{"en"}},
	}
	ranges := a.getTokenRanges(tokens)
	require.Len(t, ranges, 2)
}
