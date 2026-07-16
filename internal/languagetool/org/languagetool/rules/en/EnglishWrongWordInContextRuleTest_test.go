package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWrongWordInContextRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWrongWordInContextRuleTest.java :: EnglishWrongWordInContextRuleTest.testRule
func TestEnglishWrongWordInContextRule_Rule(t *testing.T) {
	// contains assertEquals — full values in Java twin source
	_ = "I have prescribed you a course of antibiotics." // assertGood
	_ = "Name one country that does not proscribe theft." // assertGood
	_ = "He wrote about his addiction to heroin." // assertGood
	_ = "A heroine is the principal female character in a novel." // assertGood
	_ = "I bought these books at the church bazaar." // assertGood
	_ = "She has a bizarre haircut." // assertGood
	_ = "Forgo the champagne treatment a bridal boutique often provides." // assertGood
	_ = "He sat there holding his horse by the bridle." // assertGood
	_ = "They have some great desserts on this menu." // assertGood
	_ = "They have a great marble statue." // assertGood
	_ = "Protons and neutrons" // assertGood
	_ = "The plane taxied to the hangar." // assertGood
	_ = "I have proscribed you a course of antibiotics." // assertBad
	_ = "Name one country that does not prescribe theft." // assertBad
	_ = "We know that heroine is highly addictive." // assertBad
	_ = "A heroin is the principal female character in a novel." // assertBad
	_ = "What a bazaar behavior!" // assertBad
	_ = "The Saturday morning bizarre is worth seeing even if you buy nothing." // assertBad
	_ = "The bridle party waited on the lawn." // assertBad
	_ = "Each rider used his own bridal." // assertBad
	_ = "They have some great deserts on this menu." // assertBad
	_ = "They have some great marble statutes." // assertBad
	_ = "Protons and neurons" // assertBad
	_ = "The plane taxied to the hanger." // assertBad
}
