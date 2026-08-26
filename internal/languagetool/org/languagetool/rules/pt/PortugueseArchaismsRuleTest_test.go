package pt

// Outcome + field twin of org.languagetool.rules.pt.PortugueseArchaismsRule
// (no dedicated Java *Archaisms*RuleTest; bar matches PortugueseWikipediaRule /
// PortugueseRedundancyRule / PortugueseClicheRule metadata+outcome style).
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(...).getAnalyzedSentence for this rule’s surface matching.
// PortugueseArchaismsRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas.
//
// Language modules pass different paths:
//   BrazilianPortuguese → /pt/pt-BR/archaisms.txt
//   PortugalPortuguese → /pt/pt-PT/archaisms.txt
//   Base Portuguese construction is commented out (#3095); /pt/archaisms.txt exists.
//
// Java ctor had addExamplePair commented out (câmera→câmara TODO) — metadata expects
// zero incorrect examples; do not invent a live Example pair.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from official pt-PT /pt/pt-PT/archaisms.txt rows (only list with
// live replace entries: câmera=câmara, câmeras=câmaras).
func TestPortugueseArchaismsRule_Rule(t *testing.T) {
	// Java: new PortugueseArchaismsRule(messages, "/pt/pt-PT/archaisms.txt", language)
	//       (PortugalPortuguese.getRelevantRules)
	rule := NewPortugueseArchaismsRule(nil, PortugueseArchaismsPTFilePath)

	// correct sentences (already non-archaic form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("uma câmara digital"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))

	// Official pt-PT archaisms.txt rows
	checkPortugueseArchaismsReplace(t, rule, "uma câmera digital", "câmara")
	checkPortugueseArchaismsReplace(t, rule, "várias câmeras novas", "câmaras")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	// Java getMessage() HAS trailing period after $suggestions.
	matches := rule.Match(languagetool.AnalyzePlain("uma câmera digital"))
	require.Len(t, matches, 1)
	require.Equal(t, "Arcaísmo", matches[0].GetShortMessage())
	require.Equal(t,
		"\"câmera\" é um arcaísmo. É preferível dizer <suggestion>câmara</suggestion>.",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language pt
	wantID := tools.ToId(PortugueseArchaismsReplaceID+"_"+"câmera", "pt")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_ARCHAISMS_REPLACE_CÂMERA", sr.GetID())
	// Description has no $match placeholder → unchanged after replace
	require.Equal(t, "Palavras arcaicas evitáveis", sr.GetDescription())
	// "uma " = 4; "câmera" = 6 → 4..10
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())
	require.Equal(t, []string{"câmara"}, matches[0].GetSuggestedReplacements())

	// Sentence-start: capitalized match surface + capitalized suggestion list
	matches = rule.Match(languagetool.AnalyzePlain("Câmera digital nova."))
	require.Len(t, matches, 1)
	require.Equal(t, []string{"Câmara"}, matches[0].GetSuggestedReplacements())
	require.Equal(t,
		"\"Câmera\" é um arcaísmo. É preferível dizer <suggestion>câmara</suggestion>.",
		matches[0].GetMessage())
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected at sentence start")
	// toId normalizes case → same sub-rule id as lowercase surface
	wantID = tools.ToId(PortugueseArchaismsReplaceID+"_"+"Câmera", "pt")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_ARCHAISMS_REPLACE_CÂMERA", sr.GetID())
	require.Equal(t, "Palavras arcaicas evitáveis", sr.GetDescription())
	require.Equal(t, 0, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
}

// Variant path loads official BR list only (BrazilianPortuguese.getRelevantRules path).
// Official pt-BR/archaisms.txt has only comments — no replace rows.
func TestPortugueseArchaismsRule_BRVariant(t *testing.T) {
	rule := NewPortugueseArchaismsRule(nil, PortugueseArchaismsBRFilePath)
	require.Equal(t, []string{PortugueseArchaismsBRFilePath}, rule.GetFileNames())
	// PT-only key must not match on BR-only list
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("uma câmera digital"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("várias câmeras novas"))))
}

// Base path loads official /pt/archaisms.txt (comment-only; not wired in base Portuguese).
func TestPortugueseArchaismsRule_BaseVariant(t *testing.T) {
	rule := NewPortugueseArchaismsRule(nil, PortugueseArchaismsFilePath)
	require.Equal(t, []string{PortugueseArchaismsFilePath}, rule.GetFileNames())
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("uma câmera digital"))))
}

// checkPortugueseArchaismsReplace asserts one match with the given first replacement.
func checkPortugueseArchaismsReplace(t *testing.T, rule *PortugueseArchaismsRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of PortugueseArchaismsRule constructor + overrides.
func TestPortugueseArchaismsRule_Metadata(t *testing.T) {
	rule := NewPortugueseArchaismsRule(nil, PortugueseArchaismsPTFilePath)

	require.Equal(t, PortugueseArchaismsReplaceID, rule.GetID())
	require.Equal(t, "PT_ARCHAISMS_REPLACE", rule.GetID())
	require.Equal(t, "Palavras arcaicas evitáveis", rule.GetDescription())
	require.Equal(t, "Arcaísmo", rule.GetShort())
	// Java: trailing period after $suggestions
	require.Equal(t,
		"\"$match\" é um arcaísmo. É preferível dizer $suggestions.",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugueseArchaismsPTFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/pt-PT/archaisms.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, PortugueseArchaismsURL, rule.GetURL())
	require.Equal(t,
		"https://pt.wikipedia.org/wiki/Arca%C3%ADsmo",
		rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "STYLE", rule.GetCategory().GetID().String())
	// Java: setLocQualityIssueType(ITSIssueType.LocaleViolation)
	require.Equal(t, rules.ITSLocaleViolation, rule.GetLocQualityIssueType())

	// Java example pair is commented out — zero incorrect/correct examples.
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 0)
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 0)
}

// Empty path panics (Java Objects.requireNonNull(path)).
func TestPortugueseArchaismsRule_NilPathPanics(t *testing.T) {
	require.Panics(t, func() {
		NewPortugueseArchaismsRule(nil, "")
	})
}
