package pt

// Outcome + field twin of org.languagetool.rules.pt.PortugueseWordinessRule
// (no dedicated Java *Wordiness*RuleTest; bar matches PortugueseRedundancyRuleTest /
// PortugueseClicheRuleTest / EnglishPlainEnglishRuleTest metadata+outcome style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(Portuguese).getAnalyzedSentence for this rule’s surface matching.
// PortugueseWordinessRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official wordiness.txt
// multiword keys (e.g. "raramente é o caso em que acontece") tokenize the same under
// WordTokenizer as PT.
//
// Language modules pass different paths:
//   Portuguese → /pt/wordiness.txt
//   BrazilianPortuguese → /pt/pt-BR/wordiness.txt
//   PortugalPortuguese → /pt/pt-PT/wordiness.txt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official base /pt/wordiness.txt rows.
func TestPortugueseWordinessRule_Rule(t *testing.T) {
	// Java: new PortugueseWordinessRule(messages, "/pt/wordiness.txt", language)
	//       (Portuguese.getRelevantRules)
	rule := NewPortugueseWordinessRule(nil, PortugueseWordinessFilePath)

	// correct sentences (already non-wordy form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Raramente acontece isto."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Como ele fala impressiona."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))

	// Java example: Raramente é o caso em que acontece → Raramente acontece
	// Key is lowercase in file; sentence-start capitalizes match surface + suggestion list.
	checkPortugueseWordinessReplace(t, rule, "Raramente é o caso em que acontece isto.", "Raramente acontece")

	// Official base wordiness.txt multiword rows (supplemental; not invent)
	checkPortugueseWordinessReplace(t, rule, "Vimos a forma como ele fala.", "como")
	checkPortugueseWordinessReplace(t, rule, "Foi um abrigo seguro.", "refúgio")
	checkPortugueseWordinessReplace(t, rule, "Assinaram um acordo mútuo.", "acordo")
	checkPortugueseWordinessReplace(t, rule, "Isso é absolutamente essencial.", "essencial")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	// Note: Java getMessage() HAS trailing period after $suggestions
	matches := rule.Match(languagetool.AnalyzePlain("Raramente é o caso em que acontece isto."))
	require.Len(t, matches, 1)
	require.Equal(t, "Expressão prolixa", matches[0].GetShortMessage())
	require.Equal(t,
		"\"Raramente é o caso em que acontece\" é uma expressão prolixa. É preferível dizer <suggestion>raramente acontece</suggestion>.",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language pt
	wantID := tools.ToId(PortugueseWordinessReplaceID+"_"+"Raramente é o caso em que acontece", "pt")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_WORDINESS_REPLACE_RARAMENTE_É_O_CASO_EM_QUE_ACONTECE", sr.GetID())
	require.Equal(t, "2. Expressões prolixas: Raramente é o caso em que acontece", sr.GetDescription())
	// sentence starts at 0; "Raramente é o caso em que acontece" = 34 runes (é is one code unit in BMP)
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 34, matches[0].GetToPos())
	require.Equal(t, []string{"Raramente acontece"}, matches[0].GetSuggestedReplacements())

	// Mid-sentence non-capitalized match surface
	matches = rule.Match(languagetool.AnalyzePlain("Vimos a forma como ele fala."))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"como"}, matches[0].GetSuggestedReplacements())
	require.Equal(t,
		"\"a forma como\" é uma expressão prolixa. É preferível dizer <suggestion>como</suggestion>.",
		matches[0].GetMessage())
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected mid-sentence")
	wantID = tools.ToId(PortugueseWordinessReplaceID+"_"+"a forma como", "pt")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_WORDINESS_REPLACE_A_FORMA_COMO", sr.GetID())
	require.Equal(t, "2. Expressões prolixas: a forma como", sr.GetDescription())
	// "Vimos " = 6; "a forma como" = 12 → 6..18
	require.Equal(t, 6, matches[0].GetFromPos())
	require.Equal(t, 18, matches[0].GetToPos())

	// Sentence-start capitalization of multiword suggestion (style_rules prior bar)
	matches = rule.Match(languagetool.AnalyzePlain("A forma como ele fala impressiona."))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"Como"}, matches[0].GetSuggestedReplacements())
	require.Equal(t,
		"\"A forma como\" é uma expressão prolixa. É preferível dizer <suggestion>como</suggestion>.",
		matches[0].GetMessage())
}

// Variant path loads official BR list only (BrazilianPortuguese.getRelevantRules path).
func TestPortugueseWordinessRule_BRVariant(t *testing.T) {
	rule := NewPortugueseWordinessRule(nil, PortugueseWordinessBRFilePath)
	require.Equal(t, []string{PortugueseWordinessBRFilePath}, rule.GetFileNames())
	// Official pt-BR/wordiness.txt: ativo valioso=ativo
	checkPortugueseWordinessReplace(t, rule, "Era um ativo valioso.", "ativo")
	// Base-list key is not in the BR-only file
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Vimos a forma como ele fala."))))
}

// Variant path loads official PT list only (PortugalPortuguese.getRelevantRules path).
func TestPortugueseWordinessRule_PTVariant(t *testing.T) {
	rule := NewPortugueseWordinessRule(nil, PortugueseWordinessPTFilePath)
	require.Equal(t, []string{PortugueseWordinessPTFilePath}, rule.GetFileNames())
	// Official pt-PT/wordiness.txt: activo valioso=activo
	checkPortugueseWordinessReplace(t, rule, "Era um activo valioso.", "activo")
	// Base-list key is not in the PT-only file
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Vimos a forma como ele fala."))))
}

// checkPortugueseWordinessReplace asserts one match with the given first replacement.
func checkPortugueseWordinessReplace(t *testing.T, rule *PortugueseWordinessRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of PortugueseWordinessRule constructor + overrides.
func TestPortugueseWordinessRule_Metadata(t *testing.T) {
	rule := NewPortugueseWordinessRule(nil, PortugueseWordinessFilePath)

	require.Equal(t, PortugueseWordinessReplaceID, rule.GetID())
	require.Equal(t, "PT_WORDINESS_REPLACE", rule.GetID())
	require.Equal(t, "2. Expressões prolixas: $match", rule.GetDescription())
	require.Equal(t, "Expressão prolixa", rule.GetShort())
	// Java: WITH trailing period after $suggestions (unlike Redundancy)
	require.Equal(t,
		"\"$match\" é uma expressão prolixa. É preferível dizer $suggestions.",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugueseWordinessFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/wordiness.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, PortugueseWordinessURL, rule.GetURL())
	require.Equal(t,
		"https://pt.wikipedia.org/wiki/V%C3%ADcio_de_linguagem#Prolixidade_ou_preciosismo",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REDUNDANCY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>Raramente é o caso em que acontece</marker> isto.", inc[0].GetExample())
	require.Equal(t, []string{"Raramente acontece"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>Raramente acontece</marker> isto.", corr[0].GetExample())
}
