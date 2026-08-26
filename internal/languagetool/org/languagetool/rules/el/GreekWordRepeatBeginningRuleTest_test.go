package el

// Twin of languagetool-language-modules/el/src/test/java/org/languagetool/rules/el/GreekWordRepeatBeginningRuleTest.java
//
// Java setup:
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("el"));
//   lt.check(text) → full analysis + all rules
//
// Go analysis path for this text-level rule:
//   NewJLanguageTool("el").Analyze (SRX sentence tokenize + Greek word tokenizer)
//   rule.MatchList(sents) isolates GREEK_WORD_REPEAT_BEGINNING_RULE outcomes
//   (other rules would pollute full Check match counts).
//
// Messages: MessagesBundle_el.properties keys used by WordRepeatBeginningRule
// (no invent soft strings).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// elWRBMessages ports MessagesBundle_el.properties keys used by WordRepeatBeginningRule.
func elWRBMessages() map[string]string {
	return map[string]string{
		// MessagesBundle_el desc_repetition_beginning*
		"desc_repetition_beginning":           "Συνεχόμενες προτάσεις που ξεκινάνε με την ίδια λέξη",
		"desc_repetition_beginning_word":      "Τρεις συνεχόμενες προτάσεις ξεκινάνε με την ίδια λέξη.",
		"desc_repetition_beginning_adv":       "Δύο συνεχόμενες προτάσεις ξεκινάνε με το ίδιο επίρρημα.",
		"desc_repetition_beginning_thesaurus": "Αλλάξτε την σειρά των λέξεων στην πρόταση ή συμβουλευτείτε έναν θησαυρό για να βρείτε μία συνώνυμη λέξη.",
		// REPETITIONS_STYLE message key is Java typo cateogry_repetitions_style; el has category_repetitions
		"cateogry_repetitions_style": "Επαναλήψεις",
		"category_repetitions":       "Επαναλήψεις",
	}
}

func hasSuggestion(suggs []string, want string) bool {
	for _, s := range suggs {
		if s == want {
			return true
		}
	}
	return false
}

func elWRBAtrSurface(tok string) *languagetool.AnalyzedTokenReadings {
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(tok, nil, nil))
}

// analyzeELWRBText ports JLanguageTool(el) sentence tokenize + getAnalyzedSentence per part.
func analyzeELWRBText(t *testing.T, text string) []*languagetool.AnalyzedSentence {
	t.Helper()
	lt := languagetool.NewJLanguageTool("el")
	sents := lt.Analyze(text)
	require.NotEmpty(t, sents, "Analyze(%q)", text)
	return sents
}

// TestGreekWordRepeatBeginningRule_Rule twins Java GreekWordRepeatBeginningRuleTest.testRule.
func TestGreekWordRepeatBeginningRule_Rule(t *testing.T) {
	rule := NewGreekWordRepeatBeginningRule(elWRBMessages())

	// ================== correct sentences ===================
	// two successive sentences that start with the same non-adverb word.
	require.Equal(t, 0, len(rule.MatchList(analyzeELWRBText(t, "Εγώ παίζω ποδόσφαιρο. Εγώ παίζω μπάσκετ"))))
	// three successive sentences that start with the same exception word ("Το").
	require.Equal(t, 0, len(rule.MatchList(analyzeELWRBText(t, "Το αυτοκίνητο είναι καινούργιο. Το ποδήλατο είναι παλιό. Το καράβι είναι καινούργιο."))))
	// two successive sentences that start with the same exception word ("Μία").
	require.Equal(t, 0, len(rule.MatchList(analyzeELWRBText(t, "Μία περίπτωση εξηγήθηκε ήδη. Μία άλλη θα αναλυθεί παρακάτω."))))

	// =================== errors =============================
	// two successive sentences that start with one of the saved adverbs ("Επίσης").
	matches2 := rule.MatchList(analyzeELWRBText(t, "Επίσης, μιλάω Ελληνικά. Επίσης, μιλάω Αγγλικά."))
	require.Equal(t, 1, len(matches2))
	// check suggestions (because the adverbs are contained in a Set it is safer to check if the correct suggestions
	// are contained in the real suggestions) — mirrors Java anyMatch
	suggs := matches2[0].GetSuggestedReplacements()
	require.True(t, hasSuggestion(suggs, "Επιπλέον"))
	require.True(t, hasSuggestion(suggs, "Ακόμη"))
	require.True(t, hasSuggestion(suggs, "Επιπρόσθετα"))
	require.True(t, hasSuggestion(suggs, "Συμπληρωματικά"))
	// "Επίσης" itself must not be suggested
	require.False(t, hasSuggestion(suggs, "Επίσης"))
	// adverb short message (MessagesBundle_el)
	require.Equal(t, elWRBMessages()["desc_repetition_beginning_adv"], matches2[0].GetShortMessage())
	require.Contains(t, matches2[0].GetMessage(), elWRBMessages()["desc_repetition_beginning_thesaurus"])
	// span is UTF-16 length of "Επίσης" (all BMP → 6)
	require.Equal(t, 6, matches2[0].GetToPos()-matches2[0].GetFromPos())

	// three successive sentences that start with the same non-exception word (no suggestions to check).
	matches1 := rule.MatchList(analyzeELWRBText(t, "Εγώ παίζω μπάσκετ. Εγώ παίζω ποδόσφαιρο. Εγώ παίζω βόλεϊ."))
	require.Equal(t, 1, len(matches1))
	require.Equal(t, elWRBMessages()["desc_repetition_beginning_word"], matches1[0].GetShortMessage())
	require.Contains(t, matches1[0].GetMessage(), elWRBMessages()["desc_repetition_beginning_thesaurus"])
	// empty suggestions for non-adverb three-word repeat
	require.Empty(t, matches1[0].GetSuggestedReplacements())
	// "Εγώ" UTF-16 length (BMP, combining not present → 3)
	require.Equal(t, 3, matches1[0].GetToPos()-matches1[0].GetFromPos())
}

// TestGreekWordRepeatBeginningRule_Metadata covers Java ctor id + example pair + category.
func TestGreekWordRepeatBeginningRule_Metadata(t *testing.T) {
	rule := NewGreekWordRepeatBeginningRule(elWRBMessages())
	require.Equal(t, "GREEK_WORD_REPEAT_BEGINNING_RULE", rule.GetID())
	require.Equal(t, 2, rule.MinToCheckParagraph())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), rule.GetCategory().GetID())
	require.Equal(t, elWRBMessages()["desc_repetition_beginning"], rule.GetDescription())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Επίσης, παίζω ποδόσφαιρο. <marker>Επίσης</marker>, παίζω μπάσκετ.", inc[0].GetExample())
	require.Equal(t, []string{"Επιπλέον"}, inc[0].GetCorrections())
	require.Equal(t, "Επίσης, παίζω ποδόσφαιρο. <marker>Επιπλέον</marker>, παίζω μπάσκετ.", rule.GetCorrectExamples()[0].GetExample())
}

// TestGreekWordRepeatBeginningRule_AdverbCategories ports isAdverb + getSuggestions + isException hooks.
func TestGreekWordRepeatBeginningRule_AdverbCategories(t *testing.T) {
	r := NewGreekWordRepeatBeginningRule(elWRBMessages())

	// isAdverb: ADD / CONTRAST / EXPLAIN surfaces only
	require.True(t, r.isAdverb(elWRBAtrSurface("Επίσης")))
	require.True(t, r.isAdverb(elWRBAtrSurface("Αντίθετα")))
	require.True(t, r.isAdverb(elWRBAtrSurface("Δηλαδή")))
	require.False(t, r.isAdverb(elWRBAtrSurface("Εγώ")))
	require.False(t, r.isAdverb(elWRBAtrSurface("επίσης"))) // case-sensitive HashSet
	require.False(t, r.isAdverb(nil))

	// set sizes match Java static blocks
	require.Equal(t, 5, len(elAddAdverbs))
	require.Equal(t, 4, len(elContrastAdverbs))
	require.Equal(t, 4, len(elExplainAdverbs))

	// isException hook: Greek articles only (base list OR'd in WordRepeatBeginningRule)
	for _, art := range []string{"Ο", "Η", "Το", "Οι", "Τα", "Ένας", "Μία", "Ένα"} {
		require.True(t, r.IsExceptionFn(art), "isException(%q)", art)
	}
	require.False(t, r.IsExceptionFn("Εγώ"))
	require.False(t, r.IsExceptionFn(":")) // base list not duplicated in EL hook
	// super.isException OR: base list (: – - …) still applied via WordRepeatBeginningRule.isException
	// (three successive "2011" + ":" at tokens[2] → no match, same as Java lists)
	require.Equal(t, 0, len(r.MatchList(analyzeELWRBText(t,
		"2011: ένα. 2011: δύο. 2011: τρία."))))

	// ADD suggestions: other ADD members, not self / not other sets
	s := r.getSuggestions(elWRBAtrSurface("Επίσης"))
	require.True(t, hasSuggestion(s, "Επιπλέον"))
	require.True(t, hasSuggestion(s, "Ακόμη"))
	require.True(t, hasSuggestion(s, "Επιπρόσθετα"))
	require.True(t, hasSuggestion(s, "Συμπληρωματικά"))
	require.False(t, hasSuggestion(s, "Επίσης"))
	require.False(t, hasSuggestion(s, "Αντίθετα"))
	require.Len(t, s, 4)

	// CONTRAST
	s = r.getSuggestions(elWRBAtrSurface("Ωστόσο"))
	require.True(t, hasSuggestion(s, "Αντίθετα"))
	require.True(t, hasSuggestion(s, "Εντούτοις"))
	require.True(t, hasSuggestion(s, "Εξάλλου"))
	require.False(t, hasSuggestion(s, "Ωστόσο"))
	require.Len(t, s, 3)

	// EXPLAIN
	s = r.getSuggestions(elWRBAtrSurface("Ειδικά"))
	require.True(t, hasSuggestion(s, "Δηλαδή"))
	require.True(t, hasSuggestion(s, "Ειδικότερα"))
	require.True(t, hasSuggestion(s, "Συγκεκριμένα"))
	require.False(t, hasSuggestion(s, "Ειδικά"))
	require.Len(t, s, 3)

	// non-adverb → empty
	require.Empty(t, r.getSuggestions(elWRBAtrSurface("Εγώ")))
	require.Empty(t, r.getSuggestions(nil))
}
