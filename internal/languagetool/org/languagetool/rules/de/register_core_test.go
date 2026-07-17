package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreGermanRules_Check(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-DE")
	RegisterCoreGermanRules(lt)

	require.Empty(t, lt.Check("Ein Test, der keine Fehler geben sollte."))
	// word repeat (German rule id)
	m := lt.Check("Ein Test Test, der Fehler geben sollte.")
	require.NotEmpty(t, m)
	var hasRepeat bool
	for _, x := range m {
		if x.RuleID == "GERMAN_WORD_REPEAT_RULE" || x.RuleID == "WORD_REPEAT_RULE" {
			hasRepeat = true
		}
	}
	require.True(t, hasRepeat)

	// multi whitespace
	require.NotEmpty(t, lt.Check("Hallo  Welt"))

	// double punct
	require.NotEmpty(t, lt.Check("Warte.. jetzt"))
}
