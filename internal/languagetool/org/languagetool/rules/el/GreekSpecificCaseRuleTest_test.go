package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/GreekSpecificCaseRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/GreekSpecificCaseRuleTest.java :: GreekSpecificCaseRuleTest.testRule
func TestGreekSpecificCaseRule_Rule(t *testing.T) {
	// contains assertThat
	_ = "Ηνωμένες Πολιτείες" // assertGood
	_ = "Κατοικώ στις Ηνωμένες Πολιτείες." // assertGood
	_ = "Κατοικώ στις ΗΝΩΜΕΝΕΣ ΠΟΛΙΤΕΙΕΣ." // assertGood
	_ = "ηνωμένες πολιτείες" // assertBad
	_ = "ηνωμένες Πολιτείες" // assertBad
	_ = "Ηνωμένες πολιτείες" // assertBad
	_ = "Κατοικώ στις Ηνωμένες πολιτείες." // assertBad
	_ = "Κατοικώ στις Ηνωμένες  πολιτείες." // assertBad
}
