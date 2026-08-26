package pt

// Outcome + field twin of org.languagetool.rules.pt.PortugueseRedundancyRule
// (no dedicated Java *Redundancy*RuleTest; bar matches EnglishRedundancyRuleTest /
// PortugueseClicheRuleTest / PortugueseDiacritics metadata+outcome style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(Portuguese).getAnalyzedSentence for this rule’s surface matching.
// PortugueseRedundancyRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official redundancies.txt
// multiword keys (e.g. "duna de areia") tokenize the same under WordTokenizer as PT.
//
// Language modules pass different paths:
//   Portuguese → /pt/redundancies.txt
//   BrazilianPortuguese → /pt/pt-BR/redundancies.txt
//   PortugalPortuguese → /pt/pt-PT/redundancies.txt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official base /pt/redundancies.txt rows.
func TestPortugueseRedundancyRule_Rule(t *testing.T) {
	// Java: new PortugueseRedundancyRule(messages, "/pt/redundancies.txt", language)
	//       (Portuguese.getRelevantRules)
	rule := NewPortugueseRedundancyRule(nil, PortugueseRedundancyFilePath)

	// correct sentences (already non-redundant form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A duna era enorme."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))

	// Java example: duna de areia → duna
	checkPortugueseRedundancyReplace(t, rule, "A duna de areia era enorme.", "duna")

	// Official base redundancies.txt multiword rows (supplemental; not invent)
	checkPortugueseRedundancyReplace(t, rule, "Foi um abismo profundo.", "abismo")
	checkPortugueseRedundancyReplace(t, rule, "É a absoluta verdade.", "verdade")
	checkPortugueseRedundancyReplace(t, rule, "O elo de ligação faltava.", "elo")
	checkPortugueseRedundancyReplace(t, rule, "Recebemos um alerta prévio.", "alerta")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	// Note: Java getMessage() has NO trailing period after $suggestions
	matches := rule.Match(languagetool.AnalyzePlain("A duna de areia era enorme."))
	require.Len(t, matches, 1)
	require.Equal(t, "Pleonasmo", matches[0].GetShortMessage())
	require.Equal(t,
		"\"duna de areia\" é um pleonasmo. É preferível dizer <suggestion>duna</suggestion>",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language pt
	wantID := tools.ToId(PortugueseRedundancyReplaceID+"_"+"duna de areia", "pt")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_REDUNDANCY_REPLACE_DUNA_DE_AREIA", sr.GetID())
	require.Equal(t, "1. Pleonasmos e redundâncias: duna de areia", sr.GetDescription())
	// "A " = 2; "duna de areia" = 13 → 2..15
	require.Equal(t, 2, matches[0].GetFromPos())
	require.Equal(t, 15, matches[0].GetToPos())
	require.Equal(t, []string{"duna"}, matches[0].GetSuggestedReplacements())

	// Sentence-start: capitalized match surface + capitalized suggestion list
	matches = rule.Match(languagetool.AnalyzePlain("Duna de areia era enorme."))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"Duna"}, matches[0].GetSuggestedReplacements())
	require.Equal(t,
		"\"Duna de areia\" é um pleonasmo. É preferível dizer <suggestion>duna</suggestion>",
		matches[0].GetMessage())
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected at sentence start")
	// toId normalizes case → same sub-rule id as lowercase surface
	wantID = tools.ToId(PortugueseRedundancyReplaceID+"_"+"Duna de areia", "pt")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_REDUNDANCY_REPLACE_DUNA_DE_AREIA", sr.GetID())
	require.Equal(t, "1. Pleonasmos e redundâncias: Duna de areia", sr.GetDescription())
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 13, matches[0].GetToPos())
}

// Variant path loads official BR list only (BrazilianPortuguese.getRelevantRules path).
func TestPortugueseRedundancyRule_BRVariant(t *testing.T) {
	rule := NewPortugueseRedundancyRule(nil, PortugueseRedundancyBRFilePath)
	require.Equal(t, []string{PortugueseRedundancyBRFilePath}, rule.GetFileNames())
	// Official pt-BR/redundancies.txt: labareda de fogo=labareda
	checkPortugueseRedundancyReplace(t, rule, "Vimos a labareda de fogo.", "labareda")
	// Base-list key is not in the BR-only file
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A duna de areia era enorme."))))
}

// Variant path loads official PT list only (PortugalPortuguese.getRelevantRules path).
func TestPortugueseRedundancyRule_PTVariant(t *testing.T) {
	rule := NewPortugueseRedundancyRule(nil, PortugueseRedundancyPTFilePath)
	require.Equal(t, []string{PortugueseRedundancyPTFilePath}, rule.GetFileNames())
	// Official pt-PT/redundancies.txt: bónus adicional=bónus
	checkPortugueseRedundancyReplace(t, rule, "Foi um bónus adicional.", "bónus")
	// Base-list key is not in the PT-only file
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("A duna de areia era enorme."))))
}

// checkPortugueseRedundancyReplace asserts one match with the given first replacement.
func checkPortugueseRedundancyReplace(t *testing.T, rule *PortugueseRedundancyRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of PortugueseRedundancyRule constructor + overrides.
func TestPortugueseRedundancyRule_Metadata(t *testing.T) {
	rule := NewPortugueseRedundancyRule(nil, PortugueseRedundancyFilePath)

	require.Equal(t, PortugueseRedundancyReplaceID, rule.GetID())
	require.Equal(t, "PT_REDUNDANCY_REPLACE", rule.GetID())
	require.Equal(t, "1. Pleonasmos e redundâncias: $match", rule.GetDescription())
	require.Equal(t, "Pleonasmo", rule.GetShort())
	// Java: no trailing period after $suggestions
	require.Equal(t,
		"\"$match\" é um pleonasmo. É preferível dizer $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugueseRedundancyFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/redundancies.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, PortugueseRedundancyURL, rule.GetURL())
	require.Equal(t,
		"https://pt.wikipedia.org/wiki/Pleonasmo",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "REDUNDANCY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>duna de areia</marker>", inc[0].GetExample())
	require.Equal(t, []string{"duna"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>duna</marker>", corr[0].GetExample())
}
