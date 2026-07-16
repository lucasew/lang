package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConvertToSentenceCaseFilter(t *testing.T) {
	f := NewConvertToSentenceCaseFilter()
	// "HELLO WORLD" → "Hello world" with lower lemmas
	got := f.Suggest([]SentenceCaseToken{
		{Token: "HELLO", LemmaCase: "lower"},
		{Token: "WORLD", WhitespaceBefore: true, LemmaCase: "lower"},
	})
	require.Equal(t, "Hello world", got)

	// already sentence case → suppress
	got = f.Suggest([]SentenceCaseToken{
		{Token: "Hello", LemmaCase: "lower"},
		{Token: "world", WhitespaceBefore: true, LemmaCase: "lower"},
	})
	require.Equal(t, "", got)

	// corp. abbreviation
	got = f.Suggest([]SentenceCaseToken{
		{Token: "corp", LemmaCase: "lower"},
		{Token: ".", LemmaCase: ""},
	})
	require.Equal(t, "Corp.", got)
}
