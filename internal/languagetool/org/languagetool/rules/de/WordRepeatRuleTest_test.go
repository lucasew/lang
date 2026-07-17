package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func deWordRepeat() *GermanWordRepeatRule {
	return NewGermanWordRepeatRule(map[string]string{
		"repetition": "Wiederholung",
	})
}

func assertWordRepeat(t *testing.T, text string, want int) {
	t.Helper()
	r := deWordRepeat()
	got := len(r.Match(languagetool.AnalyzePlain(text)))
	require.Equal(t, want, got, "text=%q", text)
}

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WordRepeatRuleTest.java :: WordRepeatRuleTest.testRuleGerman
func TestWordRepeatRule_RuleGerman(t *testing.T) {
	// good: relative / pronoun case exceptions
	assertWordRepeat(t, "Das sind die Sätze, die die testen sollen.", 0)
	assertWordRepeat(t, "Sätze, die die testen.", 0)
	assertWordRepeat(t, "Das Haus, auf das das Mädchen zeigt.", 0)
	assertWordRepeat(t, "Warum fragen Sie sie nicht selbst?", 0)
	assertWordRepeat(t, "Er tut das, damit sie sie nicht sieht.", 0)
	// bad: true repeats
	assertWordRepeat(t, "Die die Sätze zum testen.", 1)
	assertWordRepeat(t, "Und die die Sätze zum testen.", 1)
	assertWordRepeat(t, "Auf der der Fensterbank steht eine Blume.", 1)
	assertWordRepeat(t, "Das Buch, in in dem es steht.", 1)
	assertWordRepeat(t, "Das Haus, auf auf das Mädchen zurennen.", 1)
	assertWordRepeat(t, "Sie sie gehen nach Hause.", 1)
}
