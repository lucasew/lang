package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWordRepeatRuleTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var _ = require.Equal
var _ = tools.Unimplemented

// Port of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWordRepeatRuleTest.java :: EnglishWordRepeatRuleTest.testRepeatRule
func TestEnglishWordRepeatRule_RepeatRule(t *testing.T) {
	_ = "This is a test." // assertGood
	_ = "If I had had time, I would have gone to see him." // assertGood
	_ = "I don't think that that is a problem." // assertGood
	_ = "He also said that Azerbaijan had fulfilled a task he set, which was that that their defense budget should exceed the entire state budget of Armenia." // assertGood
	_ = "Just as if that was proof that that English was correct." // assertGood
	_ = "It was noticed after more than a month that that promise had not been carried out." // assertGood
	_ = "It was said that that lady was an actress." // assertGood
	_ = "Kurosawa's three consecutive movies after Seven Samurai had not managed to capture Japanese audiences in the way that that film had." // assertGood
	_ = "The can can hold the water." // assertGood
	_ = "May May awake up?" // assertGood
	_ = "May may awake up." // assertGood
	_ = "The cat does meow meow" // assertGood
	_ = "Hah Hah" // assertGood
	_ = "Hip Hip Hooray" // assertGood
	_ = "It's S.T.E.A.M." // assertGood
	_ = "Ok ok ok!" // assertGood
	_ = "O O O" // assertGood
	_ = "Alice and Bob had had a long-standing relationship." // assertGood
	_ = "Will Will awake up?" // assertGood
	_ = "Will will awake up." // assertGood
	_ = "Wait wait!" // assertGood
	_ = "b a s i c a l l y" // assertGood
	_ = "You can contact E.ON on Instagram." // assertGood
	_ = "In a land far far away." // assertGood
	_ = "I love you so so much." // assertGood
	_ = "Aye aye, sir!" // assertGood
	_ = "What Tom did didn't seem to bother Mary at all." // assertGood
	_ = "Whatever you do don't leave the lid up on the toilet!" // assertGood
	_ = "Keep your chin up and whatever you do don't doubt yourself or your actions." // assertGood
	_ = "I know that that can't really happen." // assertGood
	_ = "Please pass her her phone." // assertGood
	_ = "I have visited Bora Bora." // assertGood
	_ = "Hip Hip" // assertBad
	_ = "I may may awake up." // assertBad
	_ = "That is May May." // assertBad
	_ = "I will will awake up." // assertBad
	_ = "Please wait wait for me." // assertBad
	_ = "That is Will Will." // assertBad
	_ = "I will will hold the ladder." // assertBad
	_ = "You can feel confident that that this administration will continue to support a free and open Internet." // assertBad
	_ = "This is is a test." // assertBad
	_ = "But I i was not sure." // assertBad
	_ = "I I am the best." // assertBad
}
