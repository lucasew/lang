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

func TestParseAltLanguages_CommaWhitespace(t *testing.T) {
	// Java TextChecker.COMMA_WHITESPACE_PATTERN = ",\\s*"
	got := ParseAltLanguages("ru-RU, de-DE,uk-UA")
	require.Equal(t, []string{"ru-RU", "de-DE", "uk-UA"}, got)
	// plain commaSeparated keeps leading space (used for other params)
	raw := commaSeparated("ru-RU, de-DE")
	require.Equal(t, []string{"ru-RU", " de-DE"}, raw)
}

func TestForeignScriptIgnoreRanges_DiagnosticOnly(t *testing.T) {
	// Diagnostic helper still maps scripts → alt codes, but is not used for /v2/check.
	text := "Hello привет world"
	ranges := ForeignScriptIgnoreRanges(text, "en", []string{"ru-RU"})
	require.NotEmpty(t, ranges)
	require.Equal(t, "ru-RU", ranges[0].Lang)
	span := text[ranges[0].From:ranges[0].To]
	require.Contains(t, span, "привет")
}

func TestApiV2_AltLanguagesValidationAndIgnoreRanges(t *testing.T) {
	api := NewApiV2(nil, nil)
	// Bare multi-variant base rejected (Java).
	_, err := api.Handle("check", map[string]string{
		"language":     "en",
		"altLanguages": "en",
		"text":         "hi",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "variant")

	// Valid altLanguages: ignoreRanges from CheckResults (may be empty without
	// NewLanguageMatches); must not invent foreign-script ranges.
	r, err := api.Handle("check", map[string]string{
		"language":     "en",
		"altLanguages": "ru-RU, de-DE",
		"text":         "Hello привет there",
	})
	require.NoError(t, err)
	require.Contains(t, r.Body, "ignoreRanges")
	var resp CheckResponse
	require.NoError(t, json.Unmarshal([]byte(r.Body), &resp))
	// Empty is correct without speller foreign-language matches (Java path).
	require.NotNil(t, resp.IgnoreRanges)
}

func TestRangesToIgnoreRangeInfo(t *testing.T) {
	// empty stays nil/empty for JSON
	require.Empty(t, RangesToIgnoreRangeInfo(nil))
}

func TestParseCheckQueryParams_AltLanguages(t *testing.T) {
	p, err := ParseCheckQueryParams(map[string]string{
		"altLanguages": "ru-RU, de-DE",
		"language":     "en",
		"text":         "hi",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"ru-RU", "de-DE"}, p.AltLanguages)

	// Bare multi-variant rejected at parse
	_, err = ParseCheckQueryParams(map[string]string{
		"altLanguages": "en",
		"language":     "en",
		"text":         "hi",
	})
	require.Error(t, err)
}

func TestPipelineSettings_AltLanguagesInKey(t *testing.T) {
	a := pipelineSettingsFor("en", CheckOptions{AltLanguages: []string{"de-DE"}})
	b := pipelineSettingsFor("en", CheckOptions{AltLanguages: []string{"ru-RU"}})
	c := pipelineSettingsFor("en", CheckOptions{})
	require.Equal(t, []string{"de-DE"}, a.Query.AltLanguages)
	require.NotEqual(t, a.Key(), b.Key(), "different alt sets must not share pool key")
	require.NotEqual(t, a.Key(), c.Key())
}

func TestPipeline_ConfiguredLT_AltLanguages(t *testing.T) {
	settings := pipelineSettingsFor("en", CheckOptions{AltLanguages: []string{"de-DE", "ru-RU"}})
	p := NewPipeline(settings)
	lt := p.configuredLT()
	require.Equal(t, []string{"de-DE", "ru-RU"}, lt.AltLanguageCodes)
}
