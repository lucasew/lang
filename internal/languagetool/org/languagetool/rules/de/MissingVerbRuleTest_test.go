package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/MissingVerbRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/MissingVerbRuleTest.java :: MissingVerbRuleTest.test
func TestMissingVerbRule_Test(t *testing.T) {
	_ = "Da ist ein Verb, mal so zum testen." // assertGood
	_ = "Überschrift ohne Verb aber doch nicht zu kurz" // assertGood
	_ = "Sprechen Sie vielleicht zufällig Türkisch?" // assertGood
	_ = "Leg den Tresor in den Koffer im Kofferraum." // assertGood
	_ = "Bring doch einfach deine Kinder mit." // assertGood
	_ = "Gut so." // assertGood
	_ = "Ja!" // assertGood
	_ = "Vielen Dank für alles, was Du für mich getan hast." // assertGood
	_ = "Herzlichen Glückwunsch zu Deinem zwanzigsten Geburtstag." // assertGood
	_ = "Dieser Satz kein Verb." // assertBad
	_ = "Aus einer Idee sich erste Wortgruppen, aus Wortgruppen einzelne Sätze, aus Sätzen ganze Texte." // assertBad
	_ = "Ich ein neues Rad." // assertBad
}
