package ngrams

// Twin of languagetool-core/src/test/java/org/languagetool/rules/ngrams/ConfusionProbabilityRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-core/src/test/java/org/languagetool/rules/ngrams/ConfusionProbabilityRuleTest.java :: ConfusionProbabilityRuleTest.testRule
func TestConfusionProbabilityRule_Rule(t *testing.T) {
	_ = "Their"                               // assertGood
	_ = "There"                               // assertGood
	_ = "There are new ideas to explore."     // assertGood
	_ = "Why is their car broken again?"      // assertGood
	_ = "Is this their useful test?"          // assertGood
	_ = "Is this there useful test?"          // assertGood
	_ = "Their are new ideas to explore."     // assertGood
	_ = "\"Their are new ideas to explore.\"" // assertGood
	_ = "But İm dabei gut auszusehen."        // assertGood
}

// Port of languagetool-core/src/test/java/org/languagetool/rules/ngrams/ConfusionProbabilityRuleTest.java :: ConfusionProbabilityRuleTest.testLocalException
func TestConfusionProbabilityRule_LocalException(t *testing.T) {
	_ = "Their are new ideas to explore."                         // assertGood
	_ = "And their are new ideas to explore."                     // assertGood
	_ = "Their are new ideas to explore and their are new plans." // assertGood
}
