package de

import (
	"strings"
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

func TestRegisterCoreGermanRules_TextLevel(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	RegisterCoreGermanRules(lt)
	// three successive "Auch" starts
	m := lt.Check("Auch heute. Auch morgen. Auch übermorgen.")
	found := false
	for _, x := range m {
		if x.RuleID == "GERMAN_WORD_REPEAT_BEGINNING_RULE" {
			found = true
		}
	}
	require.True(t, found, "%+v", m)

	// long sentence (41+ words)
	var b strings.Builder
	for i := 0; i < 45; i++ {
		if i > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("wort")
	}
	b.WriteByte('.')
	m2 := lt.Check(b.String())
	foundLS := false
	for _, x := range m2 {
		if x.RuleID == "TOO_LONG_SENTENCE_DE" {
			foundLS = true
		}
	}
	require.True(t, foundLS, "%+v", m2)
}
