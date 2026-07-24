package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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

func TestConvertToSentenceCaseFilter_AcceptRuleMatch(t *testing.T) {
	f := NewConvertToSentenceCaseFilter()
	// HELLO WORLD inside match → Hello world
	posNN, posVB := "NN", "NN"
	t1 := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("HELLO", &posNN, strPtr("hello")),
	}, 0)
	t2 := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("WORLD", &posVB, strPtr("world")),
	}, 6)
	t2.SetWhitespaceBeforeToken(" ")
	// end positions
	// GetEndPos uses start+len
	m := NewRuleMatch(NewFakeRule("SC"), nil, 0, 11, "msg")
	out := f.AcceptRuleMatch(m, nil, 0, []*languagetool.AnalyzedTokenReadings{t1, t2}, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"Hello world"}, out.GetSuggestedReplacements())
}

func strPtr(s string) *string { return &s }
