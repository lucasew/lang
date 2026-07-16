package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/UpperCaseNgramRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/UpperCaseNgramRuleTest.java :: UpperCaseNgramRuleTest.testRule
func TestUpperCaseNgramRule_Rule(t *testing.T) {
	_ = "The New York Times reviews their gallery all the time." // assertGood
	_ = "This Was a Good Idea" // assertGood
	_ = "Professor Sprout acclimated the plant to a new environment." // assertGood
	_ = "Beauty products, Clean & Clear facial wash." // assertGood
	_ = "Please click Account > Withdraw > Update." // assertGood
	_ = "The goal is to Develop, Discuss and Learn." // assertGood
	_ = "(b) Summarize the strategy." // assertGood
	_ = "Figure/Ground:" // assertGood
	_ = "What Happened?" // assertGood
	_ = "1- Have you personally made any improvements?" // assertGood
	_ = "Lesson #1 - Create a webinar." // assertGood
	_ = "Please refund Order #5698656." // assertGood
	_ = "Let's play games at Games.co.uk." // assertGood
	_ = "Ben (Been)." // assertGood
	_ = "C stands for Curse." // assertGood
	_ = "The United States also used the short-lived slogan, \"Tastes So Good, You'll Roar\", in the early 1980s." // assertGood
	_ = "09/06 - Spoken to the business manager." // assertGood
	_ = "12.3 Game." // assertGood
	_ = "Let's talk to the Onboarding team." // assertGood
	_ = "My name is Gentle." // assertGood
	_ = "They called it Greet." // assertGood
	_ = "What is Foreshadowing?" // assertGood
	_ = "His name is Carp." // assertGood
	_ = "Victor or Rabbit as everyone calls him." // assertGood
	_ = "Think I'm Tripping?" // assertGood
	_ = "Music and Concepts." // assertGood
	_ = "It is called Ranked mode." // assertGood
	_ = "I was into Chronicle of a Death Foretold." // assertGood
	_ = "I talked with Engineering." // assertGood
	_ = "They used Draft.js to solve it." // assertGood
	_ = "And mine is Wed." // assertGood
	_ = "I would support Knicks rather than Hawks." // assertGood
	_ = "You Can't Judge a Book by the Cover" // assertGood
	_ = "What Does an Effective Cover Letter Look Like?" // assertGood
	_ = "Our external Counsel are reviewing the authority of FMPA to enter into the proposed transaction" // assertGood
	_ = "Otherwise, Staff will proceed to process your filing based on the pro forma tariff sheets submitted on August 15, 2000." // assertGood
	_ = "But he is not accomplishing enough statistically to help most Fantasy teams." // assertGood
	_ = "(4 hrs/wk) Manage all IT affairs." // assertGood
	_ = "(Laravel MVC) Implements two distinct working algorithms." // assertGood
	_ = "(Later) Connect different cont." // assertGood
	_ = "$$/month (Includes everything!)" // assertGood
	_ = "- Foot care (Cleaning of feet, wash..." // assertGood
	_ = "- Exercise (Engage in exercises..." // assertGood
	_ = "-> Allowed the civilian government..." // assertGood
	_ = "-> Led by Italian Physicians..." // assertGood
	_ = "-> Used as inspiration..." // assertGood
	_ = "The sign read \"Seats to be Added\"." // assertGood
	_ = "“Helm, Engage.”" // assertGood
	_ = "\"Be careful, Reign!\"" // assertGood
	_ = "ii) Expanded the notes." // assertGood
}

// Port of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/UpperCaseNgramRuleTest.java :: UpperCaseNgramRuleTest.testFirstLongWordToLeftIsUppercase
func TestUpperCaseNgramRule_FirstLongWordToLeftIsUppercase(t *testing.T) {
	t.Skip("unimplemented: UpperCaseNgramRuleTest.testFirstLongWordToLeftIsUppercase")
}
