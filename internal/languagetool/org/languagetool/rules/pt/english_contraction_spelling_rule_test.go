package pt

// Twin of org.languagetool.rules.pt.EnglishContractionSpellingRule
// (no dedicated Java *Test class; coverage bar matches PortugueseReplaceRule /
// EN ContractionSpellingRule ACCEPT: class fields + official list outcomes).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from official /pt/english_contractions.txt (case-sensitive).
func TestEnglishContractionSpellingRule_Rule(t *testing.T) {
	// Java: new EnglishContractionSpellingRule(messages, language)
	rule := NewEnglishContractionSpellingRule(nil)

	// correct sentences — already-formed contractions / non-keys → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ele adorava assistir what's cooking às sextas-feiras."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("I'm ill."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Don't do this."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Nada a substituir aqui."))))

	// checkSimpleReplaceRule twin: 1 match, 1 suggestion, exact message/short
	checkSimple := func(sentence, word string) {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(sentence))
		require.Equal(t, 1, len(matches), "matches for %q", sentence)
		require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()), "repl count for %q", sentence)
		require.Equal(t, word, matches[0].GetSuggestedReplacements()[0], "replacement for %q", sentence)
		// Java getMessage / getShort (match carries base AbstractSimpleReplaceRule pointer)
		require.Equal(t,
			"Caso seja uma contração da língua inglesa, prefira \""+word+"\".",
			matches[0].GetMessage(), "message for %q", sentence)
		require.Equal(t, "Erro de ortografia inglesa", matches[0].GetShortMessage(), "short for %q", sentence)
		base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
		require.True(t, ok, "rule type for %q", sentence)
		require.Equal(t, EnglishContractionSpellingRuleID, base.GetID())
	}

	// Java example surface (whats → what's)
	checkSimple("Ele adorava assistir whats cooking às sextas-feiras.", "what's")
	// Case-sensitive keys from official list
	checkSimple("Whats cooking", "What's")
	checkSimple("Dont do this", "Don't")
	checkSimple("It wasnt me", "wasn't")
	checkSimple("You neednt do this", "needn't")
	checkSimple("I know Im wrong", "I'm")
	checkSimple("youre wrong", "you're")
	checkSimple("havent earned", "haven't")
}

// Field twin of EnglishContractionSpellingRule constructor + overrides.
func TestEnglishContractionSpellingRule_Metadata(t *testing.T) {
	rule := NewEnglishContractionSpellingRule(nil)

	require.Equal(t, EnglishContractionSpellingRuleID, rule.GetID())
	require.Equal(t, "PT_ENGLISH_CONTRACTION_ORTHOGRAPHY", rule.GetID())
	require.Equal(t, "Ortografia de contrações inglesas", rule.GetDescription())
	require.Equal(t, "Erro de ortografia inglesa", rule.GetShort())
	// MessageFn with empty token/replacements (static surface for metadata probes)
	require.Equal(t,
		"Caso seja uma contração da língua inglesa, prefira \"\".",
		rule.GetMessage())

	require.Equal(t, "pt", rule.GetLocaleCode())
	require.True(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	// Java setUrl is commented out → empty URL (do not invent EN grammar-contractions URL)
	require.Equal(t, "", rule.GetURL())

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (case-sensitive keys)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Greater(t, len(ww), 0)
	require.Equal(t, []string{"what's"}, ww["whats"])
	require.Equal(t, []string{"What's"}, ww["Whats"])
	require.Equal(t, []string{"I'm"}, ww["Im"])
	// commented-out line in data: #cant=can't — not present
	_, hasCant := ww["cant"]
	require.False(t, hasCant)
	// #wont=won't — not present
	_, hasWont := ww["wont"]
	require.False(t, hasWont)

	// Category name honors supplied ResourceBundle map
	ptMsgs := map[string]string{
		"category_typo": "Erros ortográficos",
	}
	rulePT := NewEnglishContractionSpellingRule(ptMsgs)
	require.Equal(t, "Erros ortográficos", rulePT.GetCategory().GetName())
	// Hardcoded PT description/short do not depend on messages
	require.Equal(t, "Ortografia de contrações inglesas", rulePT.GetDescription())
	require.Equal(t, "Erro de ortografia inglesa", rulePT.GetShort())

	// Java addExamplePair(whats → what's)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Ele adorava assistir <marker>whats</marker> cooking às sextas-feiras.", inc[0].GetExample())
	require.Equal(t, []string{"what's"}, inc[0].GetCorrections())
	require.Equal(t, "Ele adorava assistir <marker>what's</marker> cooking às sextas-feiras.", rule.GetCorrectExamples()[0].GetExample())
}

// Load does not panic; resource parses.
func TestEnglishContractionSpellingRule_Loads(t *testing.T) {
	rule := NewEnglishContractionSpellingRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Equal(t, EnglishContractionSpellingFilePath, "/pt/english_contractions.txt")
}
