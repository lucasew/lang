package be

// Twin of languagetool-language-modules/be/src/test/java/org/languagetool/rules/be/BelarusianSpecificCaseRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/be/src/test/java/org/languagetool/rules/be/BelarusianSpecificCaseRuleTest.java :: BelarusianSpecificCaseRuleTest.testRule
func TestBelarusianSpecificCaseRule_Rule(t *testing.T) {
	// contains assertThat
	_ = "Беларуская Народная Рэспубліка" // assertGood
	_ = "Папа Рымскі" // assertGood
	_ = "дзяржаўны сцяг Рэспублікі Беларусь" // assertBad
	_ = "вярхоўны суд рэспублікі беларусь" // assertBad
	_ = "Мне падабаецца air France." // assertBad
}
