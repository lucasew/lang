package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestEnglishConvertToSentenceCaseFilter_MeException(t *testing.T) {
	f := NewEnglishConvertToSentenceCaseFilter()
	got := f.Suggest([]rules.SentenceCaseToken{
		{Token: "ME", LemmaCase: "lower"}, // exception keeps lower "me"
		{Token: "AND", WhitespaceBefore: true, LemmaCase: "lower"},
		{Token: "YOU", WhitespaceBefore: true, LemmaCase: "lower"},
	})
	// first non-punct becomes capitalized: exception still uppercased as first token → "Me"
	require.Equal(t, "Me and you", got)

	// "me" in the middle stays lower
	got = f.Suggest([]rules.SentenceCaseToken{
		{Token: "CALL", LemmaCase: "lower"},
		{Token: "ME", WhitespaceBefore: true, LemmaCase: "lower"},
		{Token: "LATER", WhitespaceBefore: true, LemmaCase: "lower"},
	})
	require.Equal(t, "Call me later", got)
}

func TestEnglishConvertToSentenceCaseFilter_Registered(t *testing.T) {
	require.True(t, patterns.GlobalRuleFilterCreator.HasFilter(
		"org.languagetool.rules.en.EnglishConvertToSentenceCaseFilter"))
	f := patterns.GlobalRuleFilterCreator.GetFilter(
		"org.languagetool.rules.en.EnglishConvertToSentenceCaseFilter")
	require.NotNil(t, f)
}
