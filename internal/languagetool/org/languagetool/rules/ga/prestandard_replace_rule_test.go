package ga

// Outcome twin of org.languagetool.rules.ga.PrestandardReplaceRule
// (no dedicated Java *RuleTest class; bar matches IrishReplaceRule ACCEPT:
// metadata + official replace-prestandard.txt outcomes + exact getMessage text).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; CheckLemmas is false so lemmas are not consulted.
// Official /ga/replace-prestandard.txt keys are surface forms (e.g. baoghal, adhbhar).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official replace-prestandard.txt rows.
func TestPrestandardReplaceRule_Rule(t *testing.T) {
	// Java: new PrestandardReplaceRule(messages, language)
	rule := NewPrestandardReplaceRule(nil)

	// correct sentences (fixed forms / non-keys) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ní baol daoibh."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("an t-ábhar sin"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jornada gan eochair prestandard."))))

	// Java example: baoghal → baol
	checkPrestandardReplace(t, rule, "Ní baoghal daoibh.", "baol")
	// full Java example sentence (curly quotes optional for surface token path)
	checkPrestandardReplace(t, rule,
		"Ní baoghal daoibh, ar sise.",
		"baol")

	// Official replace-prestandard.txt rows (surface keys as tokens)
	checkPrestandardReplace(t, rule, "an adhbhar sin", "ábhar")
	checkPrestandardReplace(t, rule, "ainim an duine", "ainm")
	checkPrestandardReplace(t, rule, "amuich as baile", "amuigh")
	// Java-visible createRuleMatch message/short for example-pair hit
	// Note: no space after colon (Java getMessage).
	matches := rule.Match(languagetool.AnalyzePlain("Ní baoghal daoibh."))
	require.Len(t, matches, 1)
	require.Equal(t, "Litriú réamhchaighdeánach", matches[0].GetShortMessage())
	require.Equal(t, "Litriú réamhchaighdeánach:\"baol\".", matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, PrestandardReplaceRuleID, base.GetID())
}

// checkPrestandardReplace asserts one match with the given first replacement and exact message.
func checkPrestandardReplace(t *testing.T, rule *PrestandardReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	// Java: "Litriú réamhchaighdeánach:\"" + join + "\"." — no space after colon
	require.Equal(t,
		"Litriú réamhchaighdeánach:\""+word+"\".",
		matches[0].GetMessage(),
		"message for %q", sentence)
	require.Equal(t, "Litriú réamhchaighdeánach", matches[0].GetShortMessage(), "short for %q", sentence)
	require.Equal(t, PrestandardReplaceRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of PrestandardReplaceRule constructor + overrides.
func TestPrestandardReplaceRule_Metadata(t *testing.T) {
	rule := NewPrestandardReplaceRule(nil)

	require.Equal(t, PrestandardReplaceRuleID, rule.GetID())
	require.Equal(t, "GA_PRESTANDARD_REPLACE", rule.GetID())
	require.Equal(t, "Litriú réamhchaighdeánach, m.sh., \"baoghal\" in áit \"baol\"", rule.GetDescription())
	require.Equal(t, "Litriú réamhchaighdeánach", rule.GetShort())
	// MessageFn with empty replacements (static surface for metadata probes)
	// No space after colon — match Java getMessage exactly.
	require.Equal(t, "Litriú réamhchaighdeánach:\"\".", rule.GetMessage())

	// Java getLocale() → Locale("GA")
	require.Equal(t, "GA", rule.GetLocaleCode())
	require.Equal(t, "GA", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.False(t, rule.SubRuleSpecificIDs)
	require.False(t, rule.IgnoreTaggedWords)
	require.Equal(t, "", rule.GetURL())

	// Java: Categories.TYPOS + ITSIssueType.Misspelling
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as /ga/replace-prestandard.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"baol"}, ww["baoghal"])
	require.Equal(t, []string{"ábhar"}, ww["adhbhar"])
	require.Equal(t, []string{"ainm"}, ww["ainim"])
	require.Equal(t, []string{"amuigh"}, ww["amuich"])
	require.Equal(t, PrestandardReplaceFilePath, "/ga/replace-prestandard.txt")

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_typo": "Botúin litrithe",
	}
	ruleMsgs := NewPrestandardReplaceRule(gaMsgs)
	require.Equal(t, "TYPOS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Botúin litrithe", ruleMsgs.GetCategory().GetName())
	// Hardcoded GA description/short do not depend on messages
	require.Equal(t, "Litriú réamhchaighdeánach, m.sh., \"baoghal\" in áit \"baol\"", ruleMsgs.GetDescription())
	require.Equal(t, "Litriú réamhchaighdeánach", ruleMsgs.GetShort())

	// Java addExamplePair(baoghal → baol)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "“Ní <marker>baoghal</marker> daoibh,” ar sise.", inc[0].GetExample())
	require.Equal(t, []string{"baol"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "“Ní <marker>baol</marker> daoibh,” ar sise.", corr[0].GetExample())
}

// Load does not panic; resource parses.
func TestPrestandardReplaceRule_Loads(t *testing.T) {
	rule := NewPrestandardReplaceRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Greater(t, len(rule.GetWrongWords()), 0)
}
