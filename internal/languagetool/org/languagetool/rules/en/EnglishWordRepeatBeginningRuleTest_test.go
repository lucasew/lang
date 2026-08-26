package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishWordRepeatBeginningRuleTest.java
//
// Java setup:
//   JLanguageTool lt = new JLanguageTool(Languages.getLanguageForShortCode("en-US"));
//   lt.check(text) → full analysis (EnglishTagger + sentence tokenize) + all rules
//
// Go analysis path for this text-level rule (same POS/surface Java uses for suggestions):
//   SRXSentenceTokenizer("en-US") + tagging/en.AnalyzeEnglishSentence per part
//   (EnglishWordTokenizer + EnglishTagger english.dict + hybrid disambiguator).
//   rule.MatchList(sents) isolates ENGLISH_WORD_REPEAT_BEGINNING_RULE outcomes.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	tagen "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func engWRBMessages() map[string]string {
	// Keys used by WordRepeatBeginningRule short/long messages.
	return map[string]string{
		"desc_repetition_beginning":           "Successive sentences beginning with the same word",
		"desc_repetition_beginning_adv":       "Three successive sentences begin with the same adverb.",
		"desc_repetition_beginning_word":      "Three successive sentences begin with the same word.",
		"desc_repetition_beginning_thesaurus": "Consider using a thesaurus to find synonyms.",
		"category_repetitions_style":          "Repetitions (Style)",
	}
}

func engWRBLT(t *testing.T) {
	t.Helper()
	if tagen.DiscoverEnglishPOSDict() == "" {
		t.Skip("english.dict not in tree (third_party/english-pos-dict or inspiration)")
	}
	tagen.EnsureDefaultEnglishTagger()
	require.NotNil(t, tagen.EnglishTagWord(), "EnglishTagWord must load from english.dict")
}

// analyzeENWRBText ports JLanguageTool(en-US).analyze/check tokenization+tagging:
// sentence tokenize then getAnalyzedSentence per part (English tagger).
func analyzeENWRBText(t *testing.T, text string) []*languagetool.AnalyzedSentence {
	t.Helper()
	parts := tokenizers.NewSRXSentenceTokenizer("en-US").Tokenize(text)
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	for _, p := range parts {
		s := tagen.AnalyzeEnglishSentence(p)
		require.NotNil(t, s, "AnalyzeEnglishSentence(%q)", p)
		out = append(out, s)
	}
	return out
}

func hasSuggestion(suggs []string, want string) bool {
	for _, s := range suggs {
		if s == want {
			return true
		}
	}
	return false
}

// TestEnglishWordRepeatBeginningRule_Rule twins Java EnglishWordRepeatBeginningRuleTest.testRule.
func TestEnglishWordRepeatBeginningRule_Rule(t *testing.T) {
	engWRBLT(t)
	rule := NewEnglishWordRepeatBeginningRule(engWRBMessages())

	// ================== correct sentences ===================
	// two successive sentences that start with the same non-adverb word.
	require.Equal(t, 0, len(rule.MatchList(analyzeENWRBText(t, "This is good. This is good, too."))))
	// three successive sentences that start with the same exception word ("the").
	require.Equal(t, 0, len(rule.MatchList(analyzeENWRBText(t, "The car. The bicycle. The third sentence with 'the'."))))

	// =================== errors =============================
	// three successive sentences that start with the same non-exception word (personal pronoun "I")
	matches1 := rule.MatchList(analyzeENWRBText(t, "I think so. I have seen that before. I don't like it."))
	require.Equal(t, 1, len(matches1))
	// check suggestions (Java ordered Arrays.asList for PRP branch)
	require.GreaterOrEqual(t, len(matches1[0].GetSuggestedReplacements()), 3)
	require.Equal(t, "Furthermore, I", matches1[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Likewise, I", matches1[0].GetSuggestedReplacements()[1])
	require.Equal(t, "Not only that, but I", matches1[0].GetSuggestedReplacements()[2])
	// word-repeat short message (three "I", not adverb)
	require.Equal(t, engWRBMessages()["desc_repetition_beginning_word"], matches1[0].GetShortMessage())
	require.Contains(t, matches1[0].GetMessage(), engWRBMessages()["desc_repetition_beginning_thesaurus"])
	// span is UTF-16 length of "I" at third sentence start
	require.Equal(t, 1, matches1[0].GetToPos()-matches1[0].GetFromPos())

	// three successive sentences that start with the same non-exception word (personal pronoun "He")
	matches2 := rule.MatchList(analyzeENWRBText(t, "He thinks so. He has seen that before. He doesn't like it."))
	require.Equal(t, 1, len(matches2))
	// check suggestions — non-"I" PRP lowercased
	require.GreaterOrEqual(t, len(matches2[0].GetSuggestedReplacements()), 3)
	require.Equal(t, "Furthermore, he", matches2[0].GetSuggestedReplacements()[0])
	require.Equal(t, "Likewise, he", matches2[0].GetSuggestedReplacements()[1])
	require.Equal(t, "Not only that, but he", matches2[0].GetSuggestedReplacements()[2])

	// two successive sentences that start with one of the saved adverbs ("Also")
	matches3 := rule.MatchList(analyzeENWRBText(t, "Also, I play football. Also, I play basketball."))
	require.Equal(t, 1, len(matches3))
	// check suggestions (because the adverbs are contained in a Set it is safer to check if the correct suggestions
	// are contained in the real suggestions) — mirrors Java anyMatch
	suggs := matches3[0].GetSuggestedReplacements()
	require.True(t, hasSuggestion(suggs, "Additionally"))
	require.True(t, hasSuggestion(suggs, "Besides"))
	require.True(t, hasSuggestion(suggs, "Furthermore"))
	require.True(t, hasSuggestion(suggs, "Moreover"))
	require.True(t, hasSuggestion(suggs, "In addition"))
	require.True(t, hasSuggestion(suggs, "As well as"))
	// "Also" itself must not be suggested (getDifferentAdverbsOfSameCategory filters self)
	require.False(t, hasSuggestion(suggs, "Also"))
	// adverb short message
	require.Equal(t, engWRBMessages()["desc_repetition_beginning_adv"], matches3[0].GetShortMessage())
	require.Equal(t, len("Also"), matches3[0].GetToPos()-matches3[0].GetFromPos())
}

// TestEnglishWordRepeatBeginningRule_Metadata covers Java constructor id + example pair + category.
func TestEnglishWordRepeatBeginningRule_Metadata(t *testing.T) {
	rule := NewEnglishWordRepeatBeginningRule(engWRBMessages())
	require.Equal(t, "ENGLISH_WORD_REPEAT_BEGINNING_RULE", rule.GetID())
	require.Equal(t, 2, rule.MinToCheckParagraph())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.NewCategoryId("REPETITIONS_STYLE"), rule.GetCategory().GetID())
	require.Equal(t, engWRBMessages()["desc_repetition_beginning"], rule.GetDescription())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Moreover, the street is almost entirely residential. <marker>Moreover</marker>, it was named after a poet.", inc[0].GetExample())
	require.Equal(t, []string{"It"}, inc[0].GetCorrections())
	require.Equal(t, "Moreover, the street is almost entirely residential. <marker>It</marker> was named after a poet.", rule.GetCorrectExamples()[0].GetExample())
}

// TestEnglishWordRepeatBeginningRule_AdverbCategories ports isAdverb + getSuggestions category arms.
func TestEnglishWordRepeatBeginningRule_AdverbCategories(t *testing.T) {
	r := NewEnglishWordRepeatBeginningRule(engWRBMessages())

	// isAdverb: ADD / CONTRAST / EMPHASIS / EXPLAIN surfaces only (not POS-based)
	require.True(t, r.isAdverb(atrSurface("Moreover")))
	require.True(t, r.isAdverb(atrSurface("Nevertheless")))
	require.True(t, r.isAdverb(atrSurface("Indeed")))
	require.True(t, r.isAdverb(atrSurface("Especially")))
	require.False(t, r.isAdverb(atrSurface("However")))
	require.False(t, r.isAdverb(nil))

	// isException hook: EN extras The/A/An (base list OR'd in WordRepeatBeginningRule)
	require.True(t, r.IsExceptionFn("The"))
	require.True(t, r.IsExceptionFn("A"))
	require.True(t, r.IsExceptionFn("An"))
	require.False(t, r.IsExceptionFn("I"))
	require.False(t, r.IsExceptionFn(":")) // base list not duplicated in EN hook

	// contrast suggestions include expressions
	s := r.getSuggestions(atrSurface("Nevertheless"))
	require.True(t, hasSuggestion(s, "Nonetheless"))
	require.True(t, hasSuggestion(s, "Alternatively"))
	require.True(t, hasSuggestion(s, "Even so"))
	require.True(t, hasSuggestion(s, "On the other hand"))
	require.False(t, hasSuggestion(s, "Nevertheless"))

	// emphasis: other emphasis only
	s = r.getSuggestions(atrSurface("Indeed"))
	require.True(t, hasSuggestion(s, "Undoubtedly"))
	require.True(t, hasSuggestion(s, "Obviously"))
	require.False(t, hasSuggestion(s, "Indeed"))
	require.False(t, hasSuggestion(s, "In addition"))

	// explain
	s = r.getSuggestions(atrSurface("Especially"))
	require.True(t, hasSuggestion(s, "Particularly"))
	require.True(t, hasSuggestion(s, "Specifically"))
	require.False(t, hasSuggestion(s, "Especially"))

	// non-PRP non-adverb → empty
	require.Empty(t, r.getSuggestions(atrSurface("This")))
}

func atrSurface(tok string) *languagetool.AnalyzedTokenReadings {
	at := languagetool.NewAnalyzedToken(tok, nil, nil)
	return languagetool.NewAnalyzedTokenReadings(at)
}
