package en

import (
	"testing"

	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/stretchr/testify/require"
)

// Twin of MorfologikAmericanSpellerRuleTest: Qur'an after full EN analysis (hybrid multiword ignore).
func TestMorfologikAmericanSpellerRule_QuranMultiword(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	sent := tagen.AnalyzeEnglishSentence("Qur'an")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms, "Java: rule.match(lt.getAnalyzedSentence(\"Qur'an\")) length 0")
}

func TestMorfologikAmericanSpellerRule_BookingComMultiword(t *testing.T) {
	r := NewMorfologikAmericanSpellerRule()
	sent := tagen.AnalyzeEnglishSentence("Booking.com is a company.")
	ms, err := r.Match(sent)
	require.NoError(t, err)
	// Booking.com multiword ignore; is/a/company need binary dict
	for _, m := range ms {
		// must not flag Booking.com tokens at start
		require.Greater(t, m.GetFromPos(), 0, "should not flag Booking.com at sentence start")
	}
}
