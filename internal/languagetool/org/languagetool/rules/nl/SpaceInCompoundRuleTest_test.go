package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/SpaceInCompoundRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/SpaceInCompoundRuleTest.java :: SpaceInCompoundRuleTest.testRule
func TestSpaceInCompoundRule_Rule(t *testing.T) {
	_ = "langeafstandloper" // assertGood
	_ = "Dat zie je nu weer met de zogenaamde oudelullendagen die in heel andere tijden met gulle hand in cao’s werden uitgereikt aan werknemers van vijftig jaar en ouder." // assertGood
	_ = "...jk aan voor de middelbare school tijdens de aanmeldw..." // assertGood
}

// Port of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/SpaceInCompoundRuleTest.java :: SpaceInCompoundRuleTest.testVariants
func TestSpaceInCompoundRule_Variants(t *testing.T) {
	// contains assertTrue
	// contains assertThat
}
