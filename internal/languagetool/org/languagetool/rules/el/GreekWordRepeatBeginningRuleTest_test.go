package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/GreekWordRepeatBeginningRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func elWRBMessages() map[string]string {
	return map[string]string{
		"desc_repetition_beginning_adv":       "Τρεις διαδοχικές προτάσεις αρχίζουν με το ίδιο επίρρημα.",
		"desc_repetition_beginning_word":      "Τρεις διαδοχικές προτάσεις αρχίζουν με την ίδια λέξη.",
		"desc_repetition_beginning_thesaurus": "Εξετάστε τη χρήση συνωνύμων.",
	}
}

func TestGreekWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewGreekWordRepeatBeginningRule(elWRBMessages())

	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Εγώ παίζω ποδόσφαιρο. Εγώ παίζω μπάσκετ"))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Το αυτοκίνητο είναι καινούργιο. Το ποδήλατο είναι παλιό. Το καράβι είναι καινούργιο."))))
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Μία περίπτωση εξηγήθηκε ήδη. Μία άλλη θα αναλυθεί παρακάτω."))))

	matches2 := rule.MatchList(languagetool.SplitAndAnalyze("Επίσης, μιλάω Ελληνικά. Επίσης, μιλάω Αγγλικά."))
	require.Equal(t, 1, len(matches2))
	suggs := matches2[0].GetSuggestedReplacements()
	has := func(s string) bool {
		for _, x := range suggs {
			if x == s {
				return true
			}
		}
		return false
	}
	require.True(t, has("Επιπλέον"))
	require.True(t, has("Ακόμη"))
	require.True(t, has("Επιπρόσθετα"))
	require.True(t, has("Συμπληρωματικά"))

	matches1 := rule.MatchList(languagetool.SplitAndAnalyze("Εγώ παίζω μπάσκετ. Εγώ παίζω ποδόσφαιρο. Εγώ παίζω βόλεϊ."))
	require.Equal(t, 1, len(matches1))
}
