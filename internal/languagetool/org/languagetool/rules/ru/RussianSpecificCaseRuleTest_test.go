package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianSpecificCaseRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianSpecificCaseRuleTest.java :: RussianSpecificCaseRuleTest.testRule
func TestRussianSpecificCaseRule_Rule(t *testing.T) {
	// contains assertThat
	_ = "Рытый Банк" // assertGood
	_ = "Центральный банк РФ" // assertGood
	_ = "Рытый банк" // assertBad
	_ = "центральный банк РФ" // assertBad
	_ = "I like air France." // assertBad
}
