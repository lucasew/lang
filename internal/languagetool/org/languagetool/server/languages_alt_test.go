package server

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultCoreLanguages_LongCode(t *testing.T) {
	langs := DefaultCoreLanguages()
	require.NotEmpty(t, langs)
	var hasEnUS, hasDeDE bool
	for _, l := range langs {
		if l.LongCode == "en-US" {
			hasEnUS = true
			require.Equal(t, "en", l.Code)
		}
		if l.LongCode == "de-DE" {
			hasDeDE = true
		}
	}
	require.True(t, hasEnUS)
	require.True(t, hasDeDE)

	api := NewApiV2(nil, nil)
	r, err := api.Handle("languages", nil)
	require.NoError(t, err)
	require.Contains(t, r.Body, "longCode")
	require.Contains(t, r.Body, "en-US")
}

func TestForeignScriptIgnoreRanges(t *testing.T) {
	text := "Hello привет world"
	ranges := ForeignScriptIgnoreRanges(text, "en", []string{"ru-RU"})
	require.NotEmpty(t, ranges)
	require.Equal(t, "ru-RU", ranges[0].Lang)
	require.Greater(t, ranges[0].To, ranges[0].From)
	// extract span
	span := text[ranges[0].From:ranges[0].To]
	require.Contains(t, span, "привет")
}

func TestApiV2_AltLanguagesIgnoreRanges(t *testing.T) {
	api := NewApiV2(nil, nil)
	r, err := api.Handle("check", map[string]string{
		"language":     "en",
		"altLanguages": "ru-RU",
		"text":         "Hello привет there",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "ignoreRanges")
	require.Contains(t, r.Body, "ru-RU")
	var resp CheckResponse
	require.NoError(t, json.Unmarshal([]byte(r.Body), &resp))
	require.NotEmpty(t, resp.IgnoreRanges)
}

func TestApiV2_AltLanguagesValidation(t *testing.T) {
	api := NewApiV2(nil, nil)
	_, err := api.Handle("check", map[string]string{
		"language":     "en",
		"altLanguages": "en",
		"text":         "hi",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "variant")
}
