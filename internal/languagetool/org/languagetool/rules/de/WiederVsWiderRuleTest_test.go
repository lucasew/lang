package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WiederVsWiderRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/WiederVsWiderRuleTest.java :: WiederVsWiderRuleTest.testRule
func TestWiederVsWiderRule_Rule(t *testing.T) {
	_ = "Das spiegelt wider, wie es wieder läuft." // assertGood
	_ = "Das spiegelt die Situation gut wider." // assertGood
	_ = "Das spiegelt die Situation." // assertGood
	_ = "Immer wieder spiegelt das die Situation." // assertGood
	_ = "Immer wieder spiegelt das die Situation wider." // assertGood
	_ = "Das spiegelt wieder wider, wie es läuft." // assertGood
	_ = "Das spiegelt wieder, wie es wieder läuft." // assertBad
	_ = "Sie spiegeln das Wachstum der Stadt wieder." // assertBad
	_ = "Das spiegelt die Situation gut wieder." // assertBad
	_ = "Immer wieder spiegelt das die Situation wieder." // assertBad
	_ = "Immer wieder spiegelte das die Situation wieder." // assertBad
}
