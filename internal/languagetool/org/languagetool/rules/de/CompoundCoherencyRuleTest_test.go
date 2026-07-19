package de

// Twin of CompoundCoherencyRuleTest (surface lemmas).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCompoundCoherencyRule_Rule(t *testing.T) {
	rule := NewCompoundCoherencyRule(nil)
	match2 := func(s1, s2 string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{
			languagetool.AnalyzePlain(s1),
			languagetool.AnalyzePlain(s2),
		}))
	}
	require.Equal(t, 0, match2("Ein Jugendfoto.", "Und ein Jugendfoto."))
	require.Equal(t, 0, match2("Der Zahn-Ärzte-Verband.", "Der Zahn-Ärzte-Verband."))

	// Jugendfoto vs Jugend-Foto
	require.Equal(t, 1, match2("Ein Jugendfoto.", "Und ein Jugend-Foto."))
	require.Equal(t, 1, match2("Ein Jugend-Foto.", "Und ein Jugendfoto."))

	// Zahn-Ärzte vs Zahnärzte
	require.Equal(t, 1, match2("Viele Zahn-Ärzte.", "Oder Zahnärzte."))

	require.Equal(t, "Einheitliche Schreibweise bei Komposita (mit oder ohne Bindestrich)", rule.GetDescription())
	require.Equal(t, -1, rule.MinToCheckParagraph())
}
