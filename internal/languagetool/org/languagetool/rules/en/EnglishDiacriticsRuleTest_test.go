package en

// Outcome twin of org.languagetool.rules.en.EnglishDiacriticsRule
// (no dedicated Java unit test; fields + AbstractSimpleReplaceRule2 match outcomes
// mirror AmericanReplaceRuleTest / BritishReplaceRuleTest style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(en).getAnalyzedSentence for this rule’s surface matching.
// EnglishDiacriticsRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official diacritics.txt
// multiword keys (e.g. "a la carte") tokenize the same under WordTokenizer as EN.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official diacritics.txt rows.
func TestEnglishDiacriticsRule_Rule(t *testing.T) {
	// Java: new EnglishDiacriticsRule(TestTools.getMessages("en"))
	//       super(messages, new English()); useSubRuleSpecificIds(); …
	rule := NewEnglishDiacriticsRule(nil)

	// correct sentences (already diacritic form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("He was quite blasé about it."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("They ordered à la carte."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nothing to replace here."))))

	// Java example: blase → blasé
	checkEnglishDiacriticsReplace(t, rule, "He was quite blase about it.", "blasé")

	// Official diacritics.txt rows
	checkEnglishDiacriticsReplace(t, rule, "They ordered a la carte.", "à la carte")
	checkEnglishDiacriticsReplace(t, rule, "His naivete was charming.", "naïveté")
	checkEnglishDiacriticsReplace(t, rule, "She prepared cafe au lait.", "café au lait")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	matches := rule.Match(languagetool.AnalyzePlain("He was quite blase about it."))
	require.Len(t, matches, 1)
	require.Equal(t, "The original word has a diacritic", matches[0].GetShortMessage())
	require.Equal(t,
		"'blase' is an imported foreign name or expression, which originally has a diacritic.",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language en
	wantID := tools.ToId(EnglishDiacriticsReplaceID+"_"+"blase", "en")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "EN_DIACRITICS_REPLACE_ORTHOGRAPHY_BLASE", sr.GetID())
	require.Equal(t, "Suggest diacritics for 'blase'", sr.GetDescription())
	require.Equal(t, 13, matches[0].GetFromPos())
	require.Equal(t, 18, matches[0].GetToPos())

	// Multiword sub-rule id (a la carte)
	matches = rule.Match(languagetool.AnalyzePlain("They ordered a la carte."))
	require.Len(t, matches, 1)
	wantID = tools.ToId(EnglishDiacriticsReplaceID+"_"+"a la carte", "en")
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected for multiword")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "EN_DIACRITICS_REPLACE_ORTHOGRAPHY_A_LA_CARTE", sr.GetID())
	require.Equal(t, "Suggest diacritics for 'a la carte'", sr.GetDescription())
	require.Equal(t, 13, matches[0].GetFromPos())
	require.Equal(t, 23, matches[0].GetToPos())
	require.Equal(t, []string{"à la carte"}, matches[0].GetSuggestedReplacements())
}

// checkEnglishDiacriticsReplace asserts one match with the given first replacement.
func checkEnglishDiacriticsReplace(t *testing.T, rule *EnglishDiacriticsRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of EnglishDiacriticsRule constructor + overrides.
func TestEnglishDiacriticsRule_Metadata(t *testing.T) {
	rule := NewEnglishDiacriticsRule(nil)

	require.Equal(t, EnglishDiacriticsReplaceID, rule.GetID())
	require.Equal(t, "EN_DIACRITICS_REPLACE_ORTHOGRAPHY", rule.GetID())
	require.Equal(t, "Suggest diacritics for '$match'", rule.GetDescription())
	require.Equal(t, "The original word has a diacritic", rule.GetShort())
	require.Equal(t,
		"'$match' is an imported foreign name or expression, which originally has a diacritic.",
		rule.GetMessage())
	require.Equal(t, " or ", rule.GetSuggestionsSeparator())
	require.Equal(t, "en", rule.GetLocaleCode())
	require.Equal(t, []string{EnglishDiacriticsFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/en/diacritics.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, EnglishDiacriticsURL, rule.GetURL())
	require.Equal(t,
		"https://en.wikipedia.org/wiki/English_terms_with_diacritical_marks",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>blase</marker>", inc[0].GetExample())
	require.Equal(t, []string{"blasé"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>blasé</marker>", corr[0].GetExample())
}
