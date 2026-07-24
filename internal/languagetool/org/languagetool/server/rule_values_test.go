package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseRuleValues(t *testing.T) {
	m := ParseRuleValues([]string{"TOO_LONG_SENTENCE:5", "OTHER:x"})
	require.Equal(t, "5", m["TOO_LONG_SENTENCE"])
	require.Equal(t, "x", m["OTHER"])
	// comma blob
	m = ParseRuleValues([]string{"A:1,B:2"})
	require.Equal(t, "1", m["A"])
	require.Equal(t, "2", m["B"])
}

func TestApiV2_RuleValuesLongSentence(t *testing.T) {
	api := NewApiV2(nil, nil)
	// short threshold so a modest sentence triggers; LongSentenceRule is Tag.picky.
	words := strings.Repeat("word ", 12)
	text := strings.TrimSpace(words) + "."
	r, err := api.Handle("check", map[string]string{
		"language":   "en",
		"level":      "picky",
		"text":       text,
		"ruleValues": "TOO_LONG_SENTENCE:5",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "TOO_LONG_SENTENCE")
	require.Contains(t, r.Body, `"id":"STYLE"`)
}

func TestApiV2_AutoDetectedLanguage(t *testing.T) {
	// language=auto sets detectedLanguage; invent preferredVariants string warning removed
	// (Java warnings object is incompleteResults only).
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language": "auto",
		"text":     "Hello world.",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "detectedLanguage")
	require.Contains(t, r.Body, "warnings")
	require.Contains(t, r.Body, "incompleteResults")
}
