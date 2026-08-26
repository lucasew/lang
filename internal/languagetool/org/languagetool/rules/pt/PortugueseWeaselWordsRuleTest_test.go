package pt

// Outcome + field twin of org.languagetool.rules.pt.PortugueseWeaselWordsRule
// (no dedicated Java *WeaselWords*RuleTest; bar matches PortugueseWordinessRule /
// PortugueseWikipediaRule / SimpleReplaceProfanityRule metadata+outcome style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(Portuguese).getAnalyzedSentence for this rule’s surface matching.
// PortugueseWeaselWordsRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official weaselwords.txt
// keys (e.g. "diz-se", "acredita-se", "a grande maioria") tokenize the same under
// WordTokenizer as PT.
//
// Java loads a fixed path only: FILE_NAME = "/pt/weaselwords.txt" (no locale variants).
// Java does NOT call useSubRuleSpecificIds().

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official /pt/weaselwords.txt rows.
func TestPortugueseWeaselWordsRule_Rule(t *testing.T) {
	// Java: new PortugueseWeaselWordsRule(messages)
	//       (commented out in Portuguese.getRelevantRules; still a public rule type)
	rule := NewPortugueseWeaselWordsRule(nil)

	// correct sentences (no weasel phrase / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Programas gratuitos podem ter qualidade."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("XYZ dizem que programas gratuitos não têm qualidade."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))

	// Java example surface: Diz-se → Quem? Como sabem?
	// Key is lowercase in file; sentence-start capitalizes match surface + suggestion list.
	checkPortugueseWeaselReplace(t, rule, "Diz-se que programas gratuitos não têm qualidade.", "Quem? Como sabem?")

	// Official base weaselwords.txt rows (supplemental; not invent)
	checkPortugueseWeaselReplace(t, rule, "Acredita-se que vai chover.", "Quem? Como sabem?")
	checkPortugueseWeaselReplace(t, rule, "Afirma-se que é verdade.", "Quem? Como sabem?")
	checkPortugueseWeaselReplace(t, rule, "Estudos mostram que isso funciona.", "Que estudos?")
	checkPortugueseWeaselReplace(t, rule, "A grande maioria concorda.", "75%? 85% 99%? Quantos?")
	checkPortugueseWeaselReplace(t, rule, "Claramente isto é falso.", "A premissa é inegavelmente verdadeira?")
	checkPortugueseWeaselReplace(t, rule, "Senso comum dita o contrário.", "Quem diz isso? O senso comum é uma prova de verdade?")

	// Java-visible createMatch extras (NO useSubRuleSpecificIds + message/short templates)
	// Java getMessage() has $match only — no $suggestions placeholder.
	matches := rule.Match(languagetool.AnalyzePlain("Diz-se que programas gratuitos não têm qualidade."))
	require.Len(t, matches, 1)
	require.Equal(t, "Expressão evasiva", matches[0].GetShortMessage())
	require.Equal(t,
		"'Diz-se' é uma expressão ambígua e evasiva. Reconsidere o seu uso, de acordo com o objetivo do seu texto.",
		matches[0].GetMessage())
	// No SpecificIdRule — useSubRuleSpecificIds is not called in Java
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	// createMatch uses the AbstractSimpleReplaceRule2 receiver (id PT_WEASELWORD_REPLACE)
	require.Equal(t, PortugueseWeaselWordReplaceID, matches[0].GetSpecificRuleId())
	require.Equal(t, "PT_WEASELWORD_REPLACE", matches[0].GetSpecificRuleId())
	// sentence starts at 0; "Diz-se" = 6 runes
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
	require.Equal(t, []string{"Quem? Como sabem?"}, matches[0].GetSuggestedReplacements())

	// Mid-sentence match surface (lowercase key form)
	matches = rule.Match(languagetool.AnalyzePlain("Aqui diz-se que chove."))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"Quem? Como sabem?"}, matches[0].GetSuggestedReplacements())
	require.Equal(t,
		"'diz-se' é uma expressão ambígua e evasiva. Reconsidere o seu uso, de acordo com o objetivo do seu texto.",
		matches[0].GetMessage())
	_, isSpecific = matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used mid-sentence")
	require.Equal(t, PortugueseWeaselWordReplaceID, matches[0].GetSpecificRuleId())
	// "Aqui " = 5; "diz-se" = 6 → 5..11
	require.Equal(t, 5, matches[0].GetFromPos())
	require.Equal(t, 11, matches[0].GetToPos())

	// Multiword mid-sentence
	matches = rule.Match(languagetool.AnalyzePlain("Segundo a grande maioria, está certo."))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"75%? 85% 99%? Quantos?"}, matches[0].GetSuggestedReplacements())
	require.Equal(t,
		"'a grande maioria' é uma expressão ambígua e evasiva. Reconsidere o seu uso, de acordo com o objetivo do seu texto.",
		matches[0].GetMessage())
	// "Segundo " = 8; "a grande maioria" = 16 → 8..24
	require.Equal(t, 8, matches[0].GetFromPos())
	require.Equal(t, 24, matches[0].GetToPos())
}

// checkPortugueseWeaselReplace asserts one match with the given first replacement.
func checkPortugueseWeaselReplace(t *testing.T, rule *PortugueseWeaselWordsRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of PortugueseWeaselWordsRule constructor + overrides.
func TestPortugueseWeaselWordsRule_Metadata(t *testing.T) {
	rule := NewPortugueseWeaselWordsRule(nil)

	require.Equal(t, PortugueseWeaselWordReplaceID, rule.GetID())
	require.Equal(t, "PT_WEASELWORD_REPLACE", rule.GetID())
	require.Equal(t, "Escrita avançada: Expressões evasivas", rule.GetDescription())
	require.Equal(t, "Expressão evasiva", rule.GetShort())
	// Java: no $suggestions in message template
	require.Equal(t,
		"'$match' é uma expressão ambígua e evasiva. Reconsidere o seu uso, de acordo com o objetivo do seu texto.",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugueseWeaselWordsFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/weaselwords.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, PortugueseWeaselWordsURL, rule.GetURL())
	require.Equal(t,
		"https://pt.wikipedia.org/wiki/Weasel_word",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>Diz-se</marker> que programas gratuitos não têm qualidade.", inc[0].GetExample())
	require.Equal(t, []string{"XYZ"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t,
		"<marker>XYZ</marker> dizem que programas gratuitos não têm qualidade. Por isso vendem programas pagos.",
		corr[0].GetExample())
}
