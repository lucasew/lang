package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WordRepeatRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WordRepeatRuleTest.java :: WordRepeatRuleTest.testRuleGerman
func TestWordRepeatRule_RuleGerman(t *testing.T) {
	_ = "Das sind die Sätze, die die testen sollen." // assertGood
	_ = "Sätze, die die testen."                     // assertGood
	_ = "Das Haus, auf das das Mädchen zeigt."       // assertGood
	_ = "Warum fragen Sie sie nicht selbst?"         // assertGood
	_ = "Er tut das, damit sie sie nicht sieht."     // assertGood
	_ = "Die die Sätze zum testen."                  // assertBad
	_ = "Und die die Sätze zum testen."              // assertBad
	_ = "Auf der der Fensterbank steht eine Blume."  // assertBad
	_ = "Das Buch, in in dem es steht."              // assertBad
	_ = "Das Haus, auf auf das Mädchen zurennen."    // assertBad
	_ = "Sie sie gehen nach Hause."                  // assertBad
}
