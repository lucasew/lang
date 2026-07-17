package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPipeline_Check(t *testing.T) {
	p := NewPipeline(NewPipelineSettings("en-US", "anon"))
	// a/an
	m := p.Check("This is an test.")
	require.NotEmpty(t, m)
	// multi space
	require.NotEmpty(t, p.Check("hello  world"))
	// disable a/an
	require.NoError(t, p.DisableRuleID("EN_A_VS_AN"))
	// still may match other rules on "an test" - only a/an disabled
	// "this this" word repeat still works
	require.NotEmpty(t, p.Check("this this"))

	// freeze blocks further disable
	p.SetupFinished()
	require.Error(t, p.DisableRuleID("WORD_REPEAT_RULE"))
}

func TestPipeline_CheckGerman(t *testing.T) {
	p := NewPipeline(NewPipelineSettings("de", "u"))
	require.NotEmpty(t, p.Check("Hallo  Welt"))
	m := p.Check("Ein Test Test.")
	require.NotEmpty(t, m)
	found := false
	for _, x := range m {
		if x.RuleID == "GERMAN_WORD_REPEAT_RULE" {
			found = true
			break
		}
	}
	require.True(t, found, "want GERMAN_WORD_REPEAT_RULE in %+v", m)
}

func TestPipeline_CheckMultiLang(t *testing.T) {
	cases := []struct {
		lang string
		text string
		id   string
	}{
		{"fr", "bonjour bonjour", "FR_WORD_REPEAT_RULE"},
		{"es", "hola hola", "SPANISH_WORD_REPEAT_RULE"},
		{"nl", "hallo hallo", "NL_WORD_REPEAT_RULE"},
		{"pl", "test test", "PL_WORD_REPEAT"},
		{"uk", "без без", "UKRAINIAN_WORD_REPEAT_RULE"},
		{"it", "ciao ciao", "ITALIAN_WORD_REPEAT_RULE"},
		{"pt", "teste teste", "PORTUGUESE_WORD_REPEAT_RULE"},
		{"ru", "тест тест", "RU_WORD_REPEAT_SIMPLE"},
		{"ca", "hola hola", "CATALAN_WORD_REPEAT_RULE"},
	}
	for _, tc := range cases {
		t.Run(tc.lang, func(t *testing.T) {
			p := NewPipeline(NewPipelineSettings(tc.lang, "u"))
			m := p.Check(tc.text)
			require.NotEmpty(t, m)
			found := false
			for _, x := range m {
				if x.RuleID == tc.id {
					found = true
					break
				}
			}
			require.True(t, found, "want %s in %+v", tc.id, m)
		})
	}
}

func TestTextChecker_CheckRemote(t *testing.T) {
	tc := NewV2TextChecker(nil, false, nil)
	ms := tc.Check("This is an test.", "en", nil)
	require.NotEmpty(t, ms)
	found := false
	for _, m := range ms {
		if m.RuleID == "EN_A_VS_AN" {
			found = true
			require.NotEmpty(t, m.Context)
			require.NotEmpty(t, m.Message)
		}
	}
	require.True(t, found)

	json, err := tc.CheckAndBuildJSON("hello  world", "en", "English", nil)
	require.NoError(t, err)
	require.Contains(t, json, "WHITESPACE_RULE")
}

func TestPipeline_EnabledOnly(t *testing.T) {
	s := NewPipelineSettings("en", "u")
	s.Query.UseEnabledOnly = true
	s.Query.EnabledRules = []string{"EN_A_VS_AN"}
	p := NewPipeline(s)
	m := p.Check("This is an test. hello  world")
	require.NotEmpty(t, m)
	for _, x := range m {
		require.Equal(t, "EN_A_VS_AN", x.RuleID)
	}
}

func TestPipeline_CheckMultiSentenceParallel(t *testing.T) {
	p := NewPipeline(NewPipelineSettings("en", "u"))
	// several sentences with a/an error
	text := "This is an test. Another line here. And an other issue."
	m := p.Check(text)
	require.NotEmpty(t, m)
	found := false
	for _, x := range m {
		if x.RuleID == "EN_A_VS_AN" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)
}
