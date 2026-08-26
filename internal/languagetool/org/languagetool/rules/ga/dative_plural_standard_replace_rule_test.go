package ga

// Outcome twin of org.languagetool.rules.ga.DativePluralStandardReplaceRule
// (no dedicated Java *RuleTest class; bar matches IrishReplaceRule ACCEPT:
// metadata + official dative-plurals simple-replacement outcomes + exact getMessage).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; CheckLemmas is false so lemmas are not consulted.
// WrongWords come from listify(DativePluralsData.getSimpleReplacements()) including
// full mutation variants (lenition / eclipse / h-prothesis).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official simple replacements.
func TestDativePluralStandardReplaceRule_Rule(t *testing.T) {
	rule := NewDativePluralStandardReplaceRule(nil)

	// correct sentences (fixed forms / non-keys) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("do na mná a gcás"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jornada gan eochair dative."))))

	// Java example: mnáibh → mná
	checkDativePlural(t, rule, "do na mnáibh a gcás", "mná")
	// full Java example sentence
	checkDativePlural(t, rule,
		"Dá mba thruamhéalach é cás an sclábhaí fir, ba mheasa fós do na mnáibh a gcás siúd.",
		"mná")

	// Official dictionary rows (surface keys as tokens)
	checkDativePlural(t, rule, "do na cosaibh sin", "cosa")
	checkDativePlural(t, rule, "ar na áiteachaibh", "áiteanna")

	// Mutation key path: lenited form maps to lenited standard
	checkDativePlural(t, rule, "do na mhnáibh", "mhná")

	// Java-visible createRuleMatch message/short for example-pair hit
	matches := rule.Match(languagetool.AnalyzePlain("do na mnáibh a gcás"))
	require.Len(t, matches, 1)
	require.Equal(t, "Tabharthach iolra", matches[0].GetShortMessage())
	require.Equal(t,
		"Is litriú réamhchaighdeánach (tabharthach iolra) é \"mnáibh\"; an \"mná\" a bhí i gceist agat?",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, DativePluralStandardReplaceRuleID, base.GetID())
}

func checkDativePlural(t *testing.T, rule *DativePluralStandardReplaceRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	// Message includes the matched token surface (tokenStr) and joined replacements.
	// Case-insensitive lookup may preserve original token casing in the message.
	require.Contains(t, matches[0].GetMessage(), "Is litriú réamhchaighdeánach (tabharthach iolra) é \"")
	require.Contains(t, matches[0].GetMessage(), "\"; an \""+word+"\" a bhí i gceist agat?")
	require.Equal(t, "Tabharthach iolra", matches[0].GetShortMessage(), "short for %q", sentence)
	require.Equal(t, DativePluralStandardReplaceRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of DativePluralStandardReplaceRule constructor + overrides.
func TestDativePluralStandardReplaceRule_Metadata(t *testing.T) {
	rule := NewDativePluralStandardReplaceRule(nil)

	require.Equal(t, DativePluralStandardReplaceRuleID, rule.GetID())
	require.Equal(t, "GA_DATIVE_PLURALS_STD", rule.GetID())
	require.Equal(t, "Tuiseal tabharthach iolra", rule.GetDescription())
	require.Equal(t, "Tabharthach iolra", rule.GetShort())
	// MessageFn with empty token/replacements (static surface for metadata probes)
	require.Equal(t,
		"Is litriú réamhchaighdeánach (tabharthach iolra) é \"\"; an \"\" a bhí i gceist agat?",
		rule.GetMessage())

	// Java: GA_LOCALE field = Locale("GA"); getLocale() not overridden (ASR default)
	require.Equal(t, "GA", rule.GetLocaleCode())
	require.Equal(t, "GA", rule.LanguageCode)
	// Java: isCaseSensitive() not overridden → AbstractSimpleReplaceRule default true
	require.True(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.False(t, rule.SubRuleSpecificIDs)
	require.False(t, rule.IgnoreTaggedWords)
	require.Equal(t, "", rule.GetURL())

	// Java: Categories.TYPOS + ITSIssueType.Grammar (not Misspelling)
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSGrammar, rule.GetLocQualityIssueType())

	// Official dictionary from DativePluralsData.getSimpleReplacements (listified)
	// Java Utils.eclipse has no 'm' case — no "mmnáibh" invent key.
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"mná"}, ww["mnáibh"])
	require.Equal(t, []string{"mhná"}, ww["mhnáibh"])
	require.Equal(t, []string{"mba"}, ww["mbúaibh"]) // eclipse of búaibh → mba
	require.Equal(t, []string{"háiteanna"}, ww["háiteachaibh"])
	require.Equal(t, []string{"háiteanna"}, ww["h-áiteachaibh"])
	require.Equal(t, DativePluralsFilePath, "/ga/dative-plurals.txt")

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_typo": "Botúin litrithe",
	}
	ruleMsgs := NewDativePluralStandardReplaceRule(gaMsgs)
	require.Equal(t, "TYPOS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Botúin litrithe", ruleMsgs.GetCategory().GetName())
	// Hardcoded GA description/short do not depend on messages
	require.Equal(t, "Tuiseal tabharthach iolra", ruleMsgs.GetDescription())
	require.Equal(t, "Tabharthach iolra", ruleMsgs.GetShort())

	// Java addExamplePair(mnáibh → mná)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t,
		"Dá mba thruamhéalach é cás an sclábhaí fir, ba mheasa fós do na <marker>mnáibh</marker> a gcás siúd.",
		inc[0].GetExample())
	require.Equal(t, []string{"mná"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t,
		"Dá mba thruamhéalach é cás an sclábhaí fir, ba mheasa fós do na <marker>mná</marker> a gcás siúd.",
		corr[0].GetExample())
}

// Load does not panic; resource parses; WrongWords is full simple map.
func TestDativePluralStandardReplaceRule_Loads(t *testing.T) {
	rule := NewDativePluralStandardReplaceRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Greater(t, len(rule.GetWrongWords()), 0)
	// Same source of truth as DativePluralsData
	simple := GetSimpleReplacements()
	require.Equal(t, len(simple), len(rule.GetWrongWords()))
	for k, v := range simple {
		require.Equal(t, []string{v}, rule.GetWrongWords()[k], "key %q", k)
	}
}
