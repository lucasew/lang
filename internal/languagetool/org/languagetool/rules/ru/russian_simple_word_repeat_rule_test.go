package ru

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// ignoreAtRU mirrors PortugueseWordRepeatRuleTest.ignore:
// analyzed tokens without whitespace, then rule.Ignore(tokens, position).
func ignoreAtRU(t *testing.T, rule *RussianSimpleWordRepeatRule, input string, position int) bool {
	t.Helper()
	sent := languagetool.AnalyzePlain(input)
	tokens := sent.GetTokensWithoutWhitespace()
	return rule.Ignore(tokens, position)
}

// matchCountRU runs WordRepeatRule.Match on AnalyzePlain sentence.
func matchCountRU(t *testing.T, rule *RussianSimpleWordRepeatRule, input string) int {
	t.Helper()
	sent := languagetool.AnalyzePlain(input)
	return len(rule.Match(sent))
}

// Behavior matrix for RussianSimpleWordRepeatRule.ignore (no dedicated Java @Test).
// Outcomes follow Java-visible wordRepetitionOf / equals / PATTERN / super.ignore.
func TestRussianSimpleWordRepeatRule_Ignore(t *testing.T) {
	rule := NewRussianSimpleWordRepeatRule(map[string]string{"repetition": "Повтор слова"})

	// wordRepetitionOf exact equals (case-sensitive)
	require.True(t, ignoreAtRU(t, rule, "и и", 2))
	require.True(t, ignoreAtRU(t, rule, "по по", 2))
	require.True(t, ignoreAtRU(t, rule, "- -", 2))
	require.True(t, ignoreAtRU(t, rule, "что что", 2))

	// Special mixed case pair only for exact "ПО"/"по" and "по"/"ПО"
	require.True(t, ignoreAtRU(t, rule, "ПО по", 2))
	require.True(t, ignoreAtRU(t, rule, "по ПО", 2))

	// Single-letter PATTERN full-string match on both, position > 1
	// Java: PATTERN.matcher(token).matches() on current and previous — any single letter.
	require.True(t, ignoreAtRU(t, rule, "L L", 2))
	require.True(t, ignoreAtRU(t, rule, "а а", 2))
	require.True(t, ignoreAtRU(t, rule, "Ё Ё", 2))
	// "И и" is NOT wordRepetitionOf("и") but BOTH match single-letter PATTERN → ignore true
	require.True(t, ignoreAtRU(t, rule, "И и", 2), "Java PATTERN single-letter arm, not wordRepetitionOf")

	// super.ignore names after ExtraIgnore returns false
	require.True(t, ignoreAtRU(t, rule, "Duran Duran", 2))
	require.True(t, ignoreAtRU(t, rule, "Wagga Wagga", 2))

	// Case-sensitive multi-letter: "По по" is NOT special ПО/по and not wordRepetitionOf("по")
	require.False(t, ignoreAtRU(t, rule, "По по", 2))
	// "ЧТО что" not exact "что"
	require.False(t, ignoreAtRU(t, rule, "ЧТО что", 2))
	// "Что что" same
	require.False(t, ignoreAtRU(t, rule, "Что что", 2))

	// Genuine multi-letter repeat should not be ignored
	require.False(t, ignoreAtRU(t, rule, "дом дом", 2))
	require.False(t, ignoreAtRU(t, rule, "no repetition", 2))
}

// Match: ignored pairs produce 0 matches; genuine repeats produce 1.
func TestRussianSimpleWordRepeatRule_Match(t *testing.T) {
	rule := NewRussianSimpleWordRepeatRule(map[string]string{
		"repetition":            "Повтор слова",
		"desc_repetition_short": "Повтор",
	})

	require.Equal(t, 0, matchCountRU(t, rule, "и и"))
	require.Equal(t, 0, matchCountRU(t, rule, "по по"))
	require.Equal(t, 0, matchCountRU(t, rule, "что что"))
	require.Equal(t, 0, matchCountRU(t, rule, "ПО по"))
	require.Equal(t, 0, matchCountRU(t, rule, "по ПО"))
	require.Equal(t, 0, matchCountRU(t, rule, "L L"))
	require.Equal(t, 0, matchCountRU(t, rule, "И и"), "single-letter PATTERN ignore")
	require.Equal(t, 0, matchCountRU(t, rule, "Duran Duran"))

	// Genuine Russian word repeat → flagged (Match uses equalsIgnoreCase then !ignore)
	require.Equal(t, 1, matchCountRU(t, rule, "дом дом"))
	// "По по": ignore false, but equalsIgnoreCase true → flagged
	require.Equal(t, 1, matchCountRU(t, rule, "По по"))
	// "Что что": multi-letter, ignore false → flagged
	require.Equal(t, 1, matchCountRU(t, rule, "Что что"))
}

func TestRussianSimpleWordRepeatRule_WordRepetitionOfCaseSensitive(t *testing.T) {
	// Direct helper parity with Java Token.equals
	tokens := []*languagetool.AnalyzedTokenReadings{
		atrRU("", false),
		atrRU("по", true),
		atrRU("по", true),
	}
	require.True(t, wordRepetitionOfRU("по", tokens, 2))
	require.False(t, wordRepetitionOfRU("ПО", tokens, 2))

	tokensMixed := []*languagetool.AnalyzedTokenReadings{
		atrRU("", false),
		atrRU("По", true),
		atrRU("по", true),
	}
	require.False(t, wordRepetitionOfRU("по", tokensMixed, 2))
	require.False(t, wordRepetitionOfRU("По", tokensMixed, 2))
}

func atrRU(surface string, wsBefore bool) *languagetool.AnalyzedTokenReadings {
	tok := languagetool.NewAnalyzedToken(surface, nil, nil)
	r := languagetool.NewAnalyzedTokenReadings(tok)
	r.SetWhitespaceBefore(wsBefore)
	return r
}

func TestRussianSimpleWordRepeatRule_ID(t *testing.T) {
	rule := NewRussianSimpleWordRepeatRule(nil)
	require.Equal(t, "WORD_REPEAT_RULE", rule.GetID())
}
