package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestEnglishConvertToSentenceCaseFilter_MeException(t *testing.T) {
	f := NewEnglishConvertToSentenceCaseFilter()
	got := f.Suggest([]rules.SentenceCaseToken{
		{Token: "ME", LemmaCase: "lower"}, // exception keeps lower "me"
		{Token: "AND", WhitespaceBefore: true, LemmaCase: "lower"},
		{Token: "YOU", WhitespaceBefore: true, LemmaCase: "lower"},
	})
	// first non-punct becomes capitalized: "me" is exception so stays "me",
	// but firstDone only after non-punct — first token is "me" (exception path
	// still goes through firstDone with UppercaseFirst of normalized "me" → "Me"
	// when TokenIsException returns true, normalized is "me", then UppercaseFirst → "Me"
	// Wait: Java returns tokenLower for exception BEFORE first capitalization.
	// Looking at Java: if exception, normalizedCase returns lowercase;
	// first token still gets uppercaseFirstChar(normalizedCase) = UppercaseFirst("me") = "Me"
	// So exception only affects non-first tokens. Verify:
	require.Equal(t, "Me and you", got)

	// "me" in the middle stays lower
	got = f.Suggest([]rules.SentenceCaseToken{
		{Token: "CALL", LemmaCase: "lower"},
		{Token: "ME", WhitespaceBefore: true, LemmaCase: "lower"},
		{Token: "LATER", WhitespaceBefore: true, LemmaCase: "lower"},
	})
	require.Equal(t, "Call me later", got)
}
