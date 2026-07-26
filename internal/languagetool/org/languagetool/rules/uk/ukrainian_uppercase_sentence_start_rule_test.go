package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Unit outcomes for UkrainianUppercaseSentenceStartRule.isException (Java override).
func TestUkrainianUppercaseSentenceStartRule_IsException(t *testing.T) {
	rule := NewUkrainianUppercaseSentenceStartRule(nil)
	require.Equal(t, "UPPERCASE_SENTENCE_START", rule.GetID())
	require.NotNil(t, rule.IsException)

	// list item: а) б) в) — tokenIdx==1, cleanToken full-matches [а-яіїєґ], next is ")"
	start := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", nil, nil),
	}, 0)
	letter := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("а", nil, nil),
	}, 0)
	paren := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(")", nil, nil),
	}, 0)
	tokens := []*languagetool.AnalyzedTokenReadings{start, letter, paren}
	require.True(t, rule.IsException(tokens, 1))

	// multi-letter cleanToken must not full-match single-letter pattern
	word := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("автор", nil, nil),
	}, 0)
	require.False(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{start, word, paren}, 1))

	// next token not ")"
	dot := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken(".", nil, nil),
	}, 0)
	require.False(t, rule.IsException([]*languagetool.AnalyzedTokenReadings{start, letter, dot}, 1))

	// tokenIdx != 1
	require.False(t, rule.IsException(tokens, 2))
}

// MatchList surface: list exception "а) …" and lowercase multi-token sentence.
func TestUkrainianUppercaseSentenceStartRule_MatchList(t *testing.T) {
	rule := NewUkrainianUppercaseSentenceStartRule(map[string]string{
		"incorrect_case": "This sentence does not start with an uppercase letter",
	})
	lt := languagetool.NewJLanguageTool("uk")
	lt.IgnoredCharacters = languagetool.UkrainianIgnoredCharactersRegex

	require.Equal(t, 0, len(rule.MatchList(lt.Analyze("а) перший пункт з кількома словами"))))
	matches := rule.MatchList(lt.Analyze("це звичайне речення з кількома словами."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "Це", matches[0].GetSuggestedReplacements()[0])
}
