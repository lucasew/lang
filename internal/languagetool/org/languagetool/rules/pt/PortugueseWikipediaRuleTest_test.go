package pt

// Outcome + field twin of org.languagetool.rules.pt.PortugueseWikipediaRule
// (no dedicated Java *Wikipedia*RuleTest; bar matches PortugueseWordinessRule /
// PortugueseRedundancyRule / PortugueseClicheRule metadata+outcome style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(Portuguese).getAnalyzedSentence for this rule’s surface matching.
// PortugueseWikipediaRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official wikipedia.txt
// multiword keys (e.g. "à diversos", "de encontro ao") tokenize the same under
// WordTokenizer as PT.
//
// Language modules pass different paths:
//   Portuguese → /pt/wikipedia.txt
//   BrazilianPortuguese → /pt/pt-BR/wikipedia.txt
//   PortugalPortuguese → /pt/pt-PT/wikipedia.txt
//
// Java ctor had addExamplePair commented out — metadata expects zero incorrect examples.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from official base /pt/wikipedia.txt rows (no invent; no Java example pair).
func TestPortugueseWikipediaRule_Rule(t *testing.T) {
	// Java: new PortugueseWikipediaRule(messages, "/pt/wikipedia.txt", language)
	//       (Portuguese.getRelevantRules)
	rule := NewPortugueseWikipediaRule(nil, PortugueseWikipediaFilePath)

	// correct sentences (already fixed form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("a diversos pedidos."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A meu ver, está certo."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))

	// Official base wikipedia.txt multiword rows
	checkPortugueseWikipediaReplace(t, rule, "Respondeu à diversos pedidos.", "a diversos")
	checkPortugueseWikipediaReplace(t, rule, "Vimos à diversas opções.", "a diversas")
	checkPortugueseWikipediaReplace(t, rule, "Foi à pedido do chefe.", "a pedido")
	checkPortugueseWikipediaReplace(t, rule, "Isso vale à todos.", "a todos")
	checkPortugueseWikipediaReplace(t, rule, "Serve à tudo.", "a tudo")
	checkPortugueseWikipediaReplace(t, rule, "Falou às outras pessoas? Não: à outras.", "a outras")
	checkPortugueseWikipediaReplace(t, rule, "Corrigiu as sua notas.", "as suas")
	checkPortugueseWikipediaReplace(t, rule, "Mora em Castanheira de Pêra.", "Castanheira de Pera")
	checkPortugueseWikipediaReplace(t, rule, "Isto foi de encontro ao esperado.", "ao encontro do")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	// Note: Java getMessage() has NO trailing period after $suggestions
	matches := rule.Match(languagetool.AnalyzePlain("Respondeu à diversos pedidos."))
	require.Len(t, matches, 1)
	require.Equal(t, "Erro gramatical ou de normativa", matches[0].GetShortMessage())
	require.Equal(t,
		"Possível erro em \"à diversos\". Prefira <suggestion>a diversos</suggestion>",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language pt
	wantID := tools.ToId(PortugueseWikipediaCommonErrorsID+"_"+"à diversos", "pt")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_WIKIPEDIA_COMMON_ERRORS_À_DIVERSOS", sr.GetID())
	require.Equal(t, "Erros frequentes nos artigos da Wikipédia: à diversos", sr.GetDescription())
	// "Respondeu " = 10; "à diversos" = 10 → 10..20
	require.Equal(t, 10, matches[0].GetFromPos())
	require.Equal(t, 20, matches[0].GetToPos())
	require.Equal(t, []string{"a diversos"}, matches[0].GetSuggestedReplacements())

	// Sentence-start capitalization of multiword suggestion
	matches = rule.Match(languagetool.AnalyzePlain("Ao meu ver, está errado."))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"A meu ver"}, matches[0].GetSuggestedReplacements())
	require.Equal(t,
		"Possível erro em \"Ao meu ver\". Prefira <suggestion>a meu ver</suggestion>",
		matches[0].GetMessage())
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected at sentence start")
	wantID = tools.ToId(PortugueseWikipediaCommonErrorsID+"_"+"Ao meu ver", "pt")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_WIKIPEDIA_COMMON_ERRORS_AO_MEU_VER", sr.GetID())
	require.Equal(t, "Erros frequentes nos artigos da Wikipédia: Ao meu ver", sr.GetDescription())
	require.Equal(t, 0, matches[0].GetFromPos())
	// "Ao meu ver" = 10 runes
	require.Equal(t, 10, matches[0].GetToPos())
}

// Variant path loads official BR list only (BrazilianPortuguese.getRelevantRules path).
func TestPortugueseWikipediaRule_BRVariant(t *testing.T) {
	rule := NewPortugueseWikipediaRule(nil, PortugueseWikipediaBRFilePath)
	require.Equal(t, []string{PortugueseWikipediaBRFilePath}, rule.GetFileNames())
	// Official pt-BR/wikipedia.txt: correiro eletronico=correio eletrônico
	checkPortugueseWikipediaReplace(t, rule, "O correiro eletronico falhou.", "correio eletrônico")
	// Official pt-BR/wikipedia.txt: peso atômico=massa atômica
	checkPortugueseWikipediaReplace(t, rule, "O peso atômico do elemento.", "massa atômica")
	// Base-list key is not in the BR-only file
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Respondeu à diversos pedidos."))))
}

// Variant path loads official PT list only (PortugalPortuguese.getRelevantRules path).
func TestPortugueseWikipediaRule_PTVariant(t *testing.T) {
	rule := NewPortugueseWikipediaRule(nil, PortugueseWikipediaPTFilePath)
	require.Equal(t, []string{PortugueseWikipediaPTFilePath}, rule.GetFileNames())
	// Official pt-PT/wikipedia.txt: correiro eletronico=correio electrónico
	checkPortugueseWikipediaReplace(t, rule, "O correiro eletronico falhou.", "correio electrónico")
	// Official pt-PT/wikipedia.txt: peso atómico=massa atómica
	checkPortugueseWikipediaReplace(t, rule, "O peso atómico do elemento.", "massa atómica")
	// Base-list key is not in the PT-only file
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Respondeu à diversos pedidos."))))
}

// checkPortugueseWikipediaReplace asserts one match with the given first replacement.
func checkPortugueseWikipediaReplace(t *testing.T, rule *PortugueseWikipediaRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of PortugueseWikipediaRule constructor + overrides.
func TestPortugueseWikipediaRule_Metadata(t *testing.T) {
	rule := NewPortugueseWikipediaRule(nil, PortugueseWikipediaFilePath)

	require.Equal(t, PortugueseWikipediaCommonErrorsID, rule.GetID())
	require.Equal(t, "PT_WIKIPEDIA_COMMON_ERRORS", rule.GetID())
	require.Equal(t, "Erros frequentes nos artigos da Wikipédia: $match", rule.GetDescription())
	require.Equal(t, "Erro gramatical ou de normativa", rule.GetShort())
	// Java: no trailing period after $suggestions
	require.Equal(t,
		"Possível erro em \"$match\". Prefira $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugueseWikipediaFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/wikipedia.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, PortugueseWikipediaURL, rule.GetURL())
	require.Equal(t,
		"https://pt.wikipedia.org/wiki/Wikip%C3%A9dia:Lista_de_erros_comuns/M%C3%A1quinas",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "WIKIPEDIA", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSGrammar, rule.GetLocQualityIssueType())

	// Java example pair is commented out — zero incorrect/correct examples.
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 0)
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 0)
}
