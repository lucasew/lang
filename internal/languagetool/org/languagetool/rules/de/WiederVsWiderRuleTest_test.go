package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WiederVsWiderRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWiederVsWiderRule_Rule(t *testing.T) {
	rule := NewWiederVsWiderRule(nil)
	ok := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), s)
	}
	bad := func(s string) {
		t.Helper()
		require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain(s))), s)
	}
	ok("Das spiegelt wider, wie es wieder läuft.")
	ok("Das spiegelt die Situation gut wider.")
	ok("Das spiegelt die Situation.")
	ok("Immer wieder spiegelt das die Situation.")
	ok("Immer wieder spiegelt das die Situation wider.")
	ok("Das spiegelt wieder wider, wie es läuft.")

	bad("Das spiegelt wieder, wie es wieder läuft.")
	bad("Sie spiegeln das Wachstum der Stadt wieder.")
	bad("Das spiegelt die Situation gut wieder.")
	bad("Immer wieder spiegelt das die Situation wieder.")
	bad("Immer wieder spiegelte das die Situation wieder.")
}
