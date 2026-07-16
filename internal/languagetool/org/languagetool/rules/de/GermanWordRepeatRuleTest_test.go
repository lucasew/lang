package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/GermanWordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanWordRepeatRule_Rule(t *testing.T) {
	rule := NewGermanWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	assertN := func(s string, n int) {
		t.Helper()
		got := len(rule.Match(languagetool.AnalyzePlain(s)))
		require.Equal(t, n, got, "sentence %q", s)
	}
	assertN("Das ist gut so.", 0)
	assertN("Das ist ist gut so.", 1)
	assertN("Der der Mann", 1)
	assertN("Warum fragen Sie sie nicht selbst?", 0)
	assertN("Er will nur sein Leben leben.", 0)
	assertN("Wie bei Honda, die die Bezahlung erhöht haben.", 0)
	assertN("Dann warfen sie sie weg.", 0)
	assertN("Dann konnte sie sie sehen.", 0)
	assertN("Er muss sein Essen essen.", 0)
	assertN("Wahrscheinlich ist das das Problem.", 0)
	assertN("Dann wäre das das erste Wirtschaftsmagazin mit mehr als 10.000 Lesern.", 0)
	assertN("Für mich war das das Härteste.", 0)
	assertN("Ich weiß, wer wer ist!", 0)
	assertN("Ich kann das machen, falls das das Problem ist.", 0)
	assertN("Als ich das das erste Mal gehört habe …", 0)
	assertN("Hat sie sie", 1)
	assertN("Hat hat", 1)
	assertN("hat hat", 1)
}
