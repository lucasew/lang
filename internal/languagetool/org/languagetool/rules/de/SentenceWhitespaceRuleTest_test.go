package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/SentenceWhitespaceRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/SentenceWhitespaceRuleTest.java :: SentenceWhitespaceRuleTest.testMatch
func TestSentenceWhitespaceRule_Match(t *testing.T) {
	// contains assertTrue
	_ = "Das ist ein Satz. Und hier der nächste." // assertGood
	_ = "Das ist ein Satz! Und hier der nächste." // assertGood
	_ = "Ist das ein Satz? Hier der nächste." // assertGood
	_ = "Am 28. September." // assertGood
	_ = "Das 1. Internationale Filmfestival findet nächste Woche statt." // assertGood
	_ = "Das ist ein Satz.Und hier der nächste." // assertBad
	_ = "Das ist ein Satz!Und hier der nächste." // assertBad
	_ = "Ist das ein Satz?Hier der nächste." // assertBad
	_ = "Am 28.September." // assertBad
	_ = "Das 1.Internationale Filmfestival findet nächste Woche statt." // assertBad
}
