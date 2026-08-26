package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WordRepeatRuleTest.java
// Java uses tagged analysis; POS / matchInflectedForms anti-patterns need tags here too.
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

func assertWordRepeatSent(t *testing.T, sent *languagetool.AnalyzedSentence, want int, label string) {
	t.Helper()
	r := deWordRepeat()
	got := len(r.Match(sent))
	require.Equal(t, want, got, "text=%q", label)
}

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WordRepeatRuleTest.java :: WordRepeatRuleTest.testRuleGerman
func TestWordRepeatRule_RuleGerman(t *testing.T) {
	// good: ", die die" via matchInflectedForms(csToken "der")
	assertWordRepeatSent(t, languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("sind", "VER:3:PLU:PRS:SFT", "sein"),
		atrWithPOS("die", "ART:DEF:NOM:PLU:MAS", "der"),
		atrWithPOS("Sätze", "SUB:NOM:PLU:MAS", "Satz"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("die", "PRO:REL:NOM:PLU:MAS", "der"),
		atrWithPOS("die", "ART:DEF:AKK:PLU:MAS", "der"),
		atrWithPOS("testen", "VER:INF:NON", "testen"),
		atrWithPOS("sollen", "VER:MOD:3:PLU:PRS:SFT", "sollen"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Das sind die Sätze, die die testen sollen.")

	assertWordRepeatSent(t, languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sätze", "SUB:NOM:PLU:MAS", "Satz"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("die", "PRO:REL:NOM:PLU:MAS", "der"),
		atrWithPOS("die", "ART:DEF:AKK:PLU:MAS", "der"),
		atrWithPOS("testen", "VER:INF:NON", "testen"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Sätze, die die testen.")

	// ", auf das das" — PRP + matchInflected der
	assertWordRepeatSent(t, languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("auf", "PRP:AKK", "auf"),
		atrWithPOS("das", "PRO:REL:AKK:SIN:NEU", "der"),
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Mädchen", "SUB:NOM:SIN:NEU", "Mädchen"),
		atrWithPOS("zeigt", "VER:3:SIN:PRS:NON", "zeigen"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Das Haus, auf das das Mädchen zeigt.")

	assertWordRepeat(t, "Warum fragen Sie sie nicht selbst?", 0)

	// damit sie sie — KON:UNT
	assertWordRepeatSent(t, languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("tut", "VER:3:SIN:PRS:NON", "tun"),
		atrWithPOS("das", "PRO:DEM:AKK:SIN:NEU", "das"),
		atrWithPOS(",", "PKT", ","),
		atrWithPOS("damit", "KON:UNT", "damit"),
		atrWithPOS("sie", "PRO:PER:NOM:SIN:FEM", "sie"),
		atrWithPOS("sie", "PRO:PER:AKK:SIN:FEM", "sie"),
		atrWithPOS("nicht", "ADV", "nicht"),
		atrWithPOS("sieht", "VER:3:SIN:PRS:NON", "sehen"),
		atrWithPOS(".", "PKT", "."),
	)), 0, "Er tut das, damit sie sie nicht sieht.")

	// bad: true repeats (untagged OK)
	assertWordRepeat(t, "Die die Sätze zum testen.", 1)
	assertWordRepeat(t, "Und die die Sätze zum testen.", 1)
	assertWordRepeat(t, "Auf der der Fensterbank steht eine Blume.", 1)
	assertWordRepeat(t, "Das Buch, in in dem es steht.", 1)
	assertWordRepeat(t, "Das Haus, auf auf das Mädchen zurennen.", 1)
	assertWordRepeat(t, "Sie sie gehen nach Hause.", 1)
}
