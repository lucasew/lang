package pt

// Twin of languagetool-language-modules/pt/src/test/java/org/languagetool/rules/pt/PortugueseClicheRuleTest.java
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(Portuguese).getAnalyzedSentence for this rule’s surface matching.
// PortugueseClicheRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official cliches.txt
// multiword keys (e.g. "a todo o vapor") tokenize the same under WordTokenizer as PT.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Twin of PortugueseClicheRuleTest.testRule
func TestPortugueseClicheRule_Rule(t *testing.T) {
	// Java: new PortugueseClicheRule(TestTools.getMessages("pt"), "/pt/cliches.txt", lt.getLanguage())
	//       lt = new JLanguageTool(Portuguese.getInstance())
	rule := NewPortugueseClicheRule(nil, PortugueseClicheFilePath)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Evite as frases-feitas e as expressões idiomáticas."))))

	// incorrect sentences:
	checkPortugueseClicheReplace(t, rule, "Teste. A todo o vapor!", "O mais rápido possível")
	checkPortugueseClicheReplace(t, rule, "A todo o vapor!", "O mais rápido possível")
	checkPortugueseClicheReplace(t, rule, "Teste, a todo o vapor!", "o mais rápido possível")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	// Mid-sentence lowercase match surface
	matches := rule.Match(languagetool.AnalyzePlain("Teste, a todo o vapor!"))
	require.Len(t, matches, 1)
	require.Equal(t, "Frase-feita", matches[0].GetShortMessage())
	require.Equal(t,
		"\"a todo o vapor\" é uma frase-feita. É preferível dizer <suggestion>o mais rápido possível</suggestion>.",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language pt
	wantID := tools.ToId(PortugueseClicheReplaceID+"_"+"a todo o vapor", "pt")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_CLICHE_REPLACE_A_TODO_O_VAPOR", sr.GetID())
	require.Equal(t, "Frases-feitas e expressões idiomáticas: a todo o vapor", sr.GetDescription())
	require.Equal(t, 7, matches[0].GetFromPos())
	require.Equal(t, 21, matches[0].GetToPos())
	require.Equal(t, []string{"o mais rápido possível"}, matches[0].GetSuggestedReplacements())

	// Sentence-start: capitalized match surface + capitalized suggestion list
	matches = rule.Match(languagetool.AnalyzePlain("A todo o vapor!"))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"O mais rápido possível"}, matches[0].GetSuggestedReplacements())
	require.Equal(t,
		"\"A todo o vapor\" é uma frase-feita. É preferível dizer <suggestion>o mais rápido possível</suggestion>.",
		matches[0].GetMessage())
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected at sentence start")
	// toId normalizes case → same sub-rule id as lowercase surface
	wantID = tools.ToId(PortugueseClicheReplaceID+"_"+"A todo o vapor", "pt")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_CLICHE_REPLACE_A_TODO_O_VAPOR", sr.GetID())
	require.Equal(t, "Frases-feitas e expressões idiomáticas: A todo o vapor", sr.GetDescription())
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 14, matches[0].GetToPos())
}

// checkPortugueseClicheReplace ports PortugueseClicheRuleTest.checkSimpleReplaceRule.
func checkPortugueseClicheReplace(t *testing.T, rule *PortugueseClicheRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of PortugueseClicheRule constructor + overrides (not all in Java test class,
// but required for sector field-by-field parity — same bar as BritishReplaceRule / Diacritics).
func TestPortugueseClicheRule_Metadata(t *testing.T) {
	rule := NewPortugueseClicheRule(nil, PortugueseClicheFilePath)

	require.Equal(t, PortugueseClicheReplaceID, rule.GetID())
	require.Equal(t, "PT_CLICHE_REPLACE", rule.GetID())
	require.Equal(t, "Frases-feitas e expressões idiomáticas: $match", rule.GetDescription())
	require.Equal(t, "Frase-feita", rule.GetShort())
	require.Equal(t,
		"\"$match\" é uma frase-feita. É preferível dizer $suggestions.",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugueseClicheFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/cliches.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, PortugueseClicheURL, rule.GetURL())
	require.Equal(t,
		"https://pt.wikipedia.org/wiki/Clichê",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>quente como uma fornalha</marker>", inc[0].GetExample())
	require.Equal(t, []string{"quente"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>quente</marker>", corr[0].GetExample())
}
