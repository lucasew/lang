package pt

// Outcome twin of org.languagetool.rules.pt.PortugueseDiacriticsRule
// (ASR2 /pt/diacritics.txt). Java PortugueseDiacriticsRuleTest.java exercises
// grammar rulegroup DIACRITICS (dialect bebé/bebê via full JLanguageTool) — not
// this ASR2 class. Do not invent a twin claiming grammar DIACRITICS is this rule.
// Coverage bar matches EN EnglishDiacriticsRule ACCEPT: class fields + official
// file match outcomes.
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface matching. PortugueseDiacriticsRule/AbstractSimpleReplaceRule2
// only inspects token surfaces, whitespaceBefore, and positions — not POS tags
// or lemmas. Official diacritics.txt multiword keys (e.g. "a la") tokenize the
// same under WordTokenizer as PT.

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official diacritics.txt rows.
func TestPortugueseDiacriticsRule_Test(t *testing.T) {
	// Java: new PortugueseDiacriticsRule(messages)
	//       super(messages, Portuguese.getInstance()); setDefaultOff(); …
	rule := NewPortugueseDiacriticsRule(nil)

	// correct sentences (already diacritic form / unrelated text):
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Um coupé clássico."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Servido à la carte."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))

	// Java example: coupe → coupé
	checkPortugueseDiacriticsReplace(t, rule, "Um coupe clássico.", "coupé")

	// Official diacritics.txt rows
	checkPortugueseDiacriticsReplace(t, rule, "Servido a la carte.", "à la carte")
	checkPortugueseDiacriticsReplace(t, rule, "Modo a la.", "à la")
	checkPortugueseDiacriticsReplace(t, rule, "Um aperitif gelado.", "apéritif")

	// Case-insensitive matching (ASR2 CI + Locale("pt"))
	checkPortugueseDiacriticsReplace(t, rule, "Um Coupe clássico.", "Coupé")

	// Java-visible createMatch extras (useSubRuleSpecificIds + message/short templates)
	matches := rule.Match(languagetool.AnalyzePlain("Um coupe clássico."))
	require.Len(t, matches, 1)
	require.Equal(t, "A palavra estrangeira original tem diacrítico", matches[0].GetShortMessage())
	require.Equal(t,
		"'coupe' é uma expressão estrangeira importada cuja grafia tem diacríticos. É preferível escrever <suggestion>coupé</suggestion>",
		matches[0].GetMessage())
	// StringTools.toId(getId() + "_" + originalStr, language) with language pt
	wantID := tools.ToId(PortugueseDiacriticsReplaceID+"_"+"coupe", "pt")
	sr, ok := matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_DIACRITICS_REPLACE_COUPE", sr.GetID())
	require.Equal(t, "Palavras estrangeiras com diacríticos: coupe", sr.GetDescription())
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 8, matches[0].GetToPos())

	// Multiword sub-rule id (a la carte)
	matches = rule.Match(languagetool.AnalyzePlain("Servido a la carte."))
	require.Len(t, matches, 1)
	wantID = tools.ToId(PortugueseDiacriticsReplaceID+"_"+"a la carte", "pt")
	sr, ok = matches[0].GetRule().(*rules.SpecificIdRule)
	require.True(t, ok, "sub-rule specific id expected for multiword")
	require.Equal(t, wantID, sr.GetID())
	require.Equal(t, "PT_DIACRITICS_REPLACE_A_LA_CARTE", sr.GetID())
	require.Equal(t, "Palavras estrangeiras com diacríticos: a la carte", sr.GetDescription())
	require.Equal(t, 8, matches[0].GetFromPos())
	require.Equal(t, 18, matches[0].GetToPos())
	require.Equal(t, []string{"à la carte"}, matches[0].GetSuggestedReplacements())
}

// checkPortugueseDiacriticsReplace asserts one match with the given first replacement.
func checkPortugueseDiacriticsReplace(t *testing.T, rule *PortugueseDiacriticsRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()),
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
}

// Field twin of PortugueseDiacriticsRule constructor + overrides.
func TestPortugueseDiacriticsRule_Metadata(t *testing.T) {
	rule := NewPortugueseDiacriticsRule(nil)

	require.Equal(t, PortugueseDiacriticsReplaceID, rule.GetID())
	require.Equal(t, "PT_DIACRITICS_REPLACE", rule.GetID())
	require.Equal(t, "Palavras estrangeiras com diacríticos: $match", rule.GetDescription())
	require.Equal(t, "A palavra estrangeira original tem diacrítico", rule.GetShort())
	require.Equal(t,
		"'$match' é uma expressão estrangeira importada cuja grafia tem diacríticos. É preferível escrever $suggestions",
		rule.GetMessage())
	require.Equal(t, " ou ", rule.GetSuggestionsSeparator())
	require.Equal(t, "pt", rule.GetLocaleCode())
	require.Equal(t, []string{PortugueseDiacriticsFilePath}, rule.GetFileNames())
	require.Equal(t, []string{"/pt/diacritics.txt"}, rule.GetFileNames())
	require.True(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	require.Equal(t, PortugueseDiacriticsURL, rule.GetURL())
	require.Equal(t,
		"http://escreverbem.com.br/o-frances-no-portugues-2/",
		rule.GetURL())
	require.True(t, rule.IsDefaultOff())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>coupe</marker>", inc[0].GetExample())
	require.Equal(t, []string{"coupé"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "<marker>coupé</marker>", corr[0].GetExample())
}
