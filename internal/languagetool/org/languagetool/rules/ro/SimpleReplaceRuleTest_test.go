package ro

// Twin of languagetool-language-modules/ro/src/test/java/org/languagetool/rules/ro/SimpleReplaceRuleTest.java
//
// Analysis path: AnalyzeWithTokenizer + RomanianWordTokenizer (WordTokenizerForLanguage("ro"))
// matches Java JLanguageTool(new Romanian()).getAnalyzedSentence for this rule’s surface
// matching. AbstractSimpleReplaceRule2 only inspects token surfaces, whitespaceBefore,
// and positions — not POS tags or lemmas. Hyphenated keys (cou-boi, cau-boi) require RO
// delimiters (hyphen is a token delimiter and is re-joined without whitespace).

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// analyzeRO mirrors Java lt.getAnalyzedSentence for SimpleReplaceRule tests
// (Romanian word tokenizer; untagged surfaces — sufficient for ASR2).
func analyzeRO(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizer(text, languagetool.WordTokenizerForLanguage("ro"))
}

// Java SimpleReplaceRuleTest.testInvalidSuggestion:
// Make sure that the suggested word is not the same as the wrong word.
func TestSimpleReplaceRule_InvalidSuggestion(t *testing.T) {
	rule := NewSimpleReplaceRule(nil)
	var invalid []string
	for _, ruleEntry := range rule.GetWrongWords() {
		for fromWord, swm := range ruleEntry {
			toWord := swm.Suggestion
			if toWord == "" || fromWord == toWord {
				invalid = append(invalid, toWord)
			}
		}
	}
	require.Empty(t, invalid, "Invalid suggestions found for: %v", invalid)
}

// Java SimpleReplaceRuleTest.testRule full matrix.
func TestSimpleReplaceRule_Rule(t *testing.T) {
	// Java: new SimpleReplaceRule(TestTools.getMessages("ro"));
	//       JLanguageTool lt = new JLanguageTool(new Romanian());
	rule := NewSimpleReplaceRule(nil)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(analyzeRO("Paisprezece case."))))

	// incorrect sentences:

	// at the beginning of a sentence (case preserved via createMatch)
	checkSimpleReplaceRule(t, rule, "Patrusprezece case.", "Paisprezece")
	// inside sentence
	checkSimpleReplaceRule(t, rule, "Satul are patrusprezece case.", "paisprezece")
	checkSimpleReplaceRule(t, rule, "Satul are (patrusprezece) case.", "paisprezece")
	checkSimpleReplaceRule(t, rule, "Satul are «patrusprezece» case.", "paisprezece")

	checkSimpleReplaceRule(t, rule, "El are șasesprezece ani.", "șaisprezece")
	checkSimpleReplaceRule(t, rule, "El a luptat pentru întâiele cărți.", "întâile")
	checkSimpleReplaceRule(t, rule, "El are cinsprezece cărți.", "cincisprezece")
	checkSimpleReplaceRule(t, rule, "El a fost patruzecioptist.", "pașoptist")
	checkSimpleReplaceRule(t, rule, "M-am adresat întâiei venite.", "întâii")
	checkSimpleReplaceRule(t, rule, "M-am adresat întâielor venite.", "întâilor")
	checkSimpleReplaceRule(t, rule, "A ajuns al douăzecelea.", "douăzecilea")
	checkSimpleReplaceRule(t, rule, "A ajuns al zecilea.", "zecelea")
	checkSimpleReplaceRule(t, rule, "A primit jumate de litru de lapte.", "jumătate")

	// multiple words / compounds
	// space-delimited
	checkSimpleReplaceRule(t, rule, "aqua forte", "Acvaforte")
	checkSimpleReplaceRule(t, rule, "aqua forte.", "Acvaforte")
	checkSimpleReplaceRule(t, rule, "A folosit «aqua forte».", "acvaforte")
	checkSimpleReplaceRule(t, rule, "Aqua forte.", "Acvaforte")
	checkSimpleReplaceRule(t, rule, "este aqua forte", "acvaforte")
	checkSimpleReplaceRule(t, rule, "este aqua forte.", "acvaforte")
	checkSimpleReplaceRule(t, rule, "este Aqua Forte.", "Acvaforte")
	checkSimpleReplaceRule(t, rule, "este AquA Forte.", "acvaforte")
	checkSimpleReplaceRule(t, rule, "A primit jumate de litru de lapte și este aqua forte.", "jumătate", "acvaforte")
	checkSimpleReplaceRule(t, rule, "du-te vino", "Du-te-vino")
	// dash-delimited
	checkSimpleReplaceRule(t, rule, "cou-boi", "Cowboy")
	checkSimpleReplaceRule(t, rule, "cow-boy", "Cowboy")
	checkSimpleReplaceRule(t, rule, "cau-boi", "Cowboy")
	checkSimpleReplaceRule(t, rule, "Cau-boi", "Cowboy")
	checkSimpleReplaceRule(t, rule, "cowboy") // correct, no replacement
	checkSimpleReplaceRule(t, rule, "Iată un cau-boi", "cowboy")
	checkSimpleReplaceRule(t, rule, "Iată un cau-boi.", "cowboy")
	checkSimpleReplaceRule(t, rule, "Iată un (cau-boi).", "cowboy")
	checkSimpleReplaceRule(t, rule, "văcar=cau-boi", "cowboy")

	// multiple suggestions
	checkSimpleReplaceRule(t, rule, "A fost adăogită o altă regulă.", "adăugită/adăugată")
	checkSimpleReplaceRule(t, rule, "A venit adinioarea.", "adineaori/adineauri")

	// words with multiple wrong forms
	checkSimpleReplaceRule(t, rule, "A pus axterix.", "asterisc")
	checkSimpleReplaceRule(t, rule, "A pus axterics.", "asterisc")
	checkSimpleReplaceRule(t, rule, "A pus asterics.", "asterisc")
}

// Field twin of SimpleReplaceRule constructor + overrides.
func TestSimpleReplaceRule_Metadata(t *testing.T) {
	// Java: TestTools.getMessages("ro") → category_misc = "Diverse"
	msgs := map[string]string{"category_misc": "Diverse"}
	rule := NewSimpleReplaceRule(msgs)

	require.Equal(t, RomanianSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "RO_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Cuvinte sau grupuri de cuvinte incorecte sau ieșite din uz", rule.GetDescription())
	require.Equal(t, "Cuvânt incorect sau ieșit din uz", rule.GetShort())
	require.Equal(t, "'$match' este incorect sau ieșit din uz, folosiți $suggestions", rule.GetMessage())
	// Java getLocale() → new Locale("ro")
	require.Equal(t, "ro", rule.GetLocaleCode())
	require.Equal(t, RomanianSimpleReplaceFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/ro/replace.txt"}, rule.GetFileNames())
	// Java getSuggestionsSeparator override
	require.Equal(t, " sau ", rule.GetSuggestionsSeparator())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	// Java AbstractSimpleReplaceRule2 default CaseSensitivy.CI; RO does not override
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	// Java: setCategory(Categories.MISC.getCategory(messages)) with MessagesBundle_ro
	require.Equal(t, "Diverse", rule.GetCategory().GetName())

	// Spot-check createMatch message / short on a single-word hit
	matches := rule.Match(analyzeRO("Satul are patrusprezece case."))
	require.Len(t, matches, 1)
	require.Equal(t, "Cuvânt incorect sau ieșit din uz", matches[0].GetShortMessage())
	require.Equal(t,
		"'patrusprezece' este incorect sau ieșit din uz, folosiți <suggestion>paisprezece</suggestion>",
		matches[0].GetMessage())
	// multi-suggestion separator " sau " in message template
	matches2 := rule.Match(analyzeRO("A venit adinioarea."))
	require.Len(t, matches2, 1)
	require.Equal(t,
		"'adinioarea' este incorect sau ieșit din uz, folosiți <suggestion>adineaori</suggestion> sau <suggestion>adineauri</suggestion>",
		matches2[0].GetMessage())
	// no sub-rule specific ids
	_, isSub := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSub)
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok)
	require.Equal(t, RomanianSimpleReplaceRuleID, base.GetID())
}

// checkSimpleReplaceRule ports Java private helper of the same name.
// words: expected suggested replacements per match; use "/" to separate multiple forms.
// Zero words → expect zero matches (correct sentence for that form).
func checkSimpleReplaceRule(t *testing.T, rule *SimpleReplaceRule, sentence string, words ...string) {
	t.Helper()
	matches := rule.Match(analyzeRO(sentence))
	require.Equal(t, len(words), len(matches), "Invalid matches.length while checking sentence: %s (got %v)", sentence, matchSummaries(matches))
	for i, word := range words {
		replacements := strings.Split(word, "/")
		got := matches[i].GetSuggestedReplacements()
		require.Equal(t, len(replacements), len(got),
			"Invalid replacement count while checking sentence: %s", sentence)
		for j, want := range replacements {
			require.Equal(t, want, got[j],
				"Invalid suggested replacement while checking sentence: %s", sentence)
		}
	}
}

func matchSummaries(matches []*rules.RuleMatch) []string {
	out := make([]string, len(matches))
	for i, m := range matches {
		out[i] = strings.Join(m.GetSuggestedReplacements(), "/")
	}
	return out
}
