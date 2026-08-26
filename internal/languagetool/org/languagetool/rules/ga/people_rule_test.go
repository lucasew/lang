package ga

// Outcome twin of org.languagetool.rules.ga.PeopleRule
// (no dedicated Java *RuleTest class; bar matches Logainm / IrishFGBEq ACCEPT:
// full metadata + Match outcomes + exact getMessage text).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; CheckLemmas is false so lemmas are not consulted.
// Official /ga/names.txt keys are English personal-name surfaces (e.g. Damocles).
// Java does not call setIgnoreTaggedWords (false).

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Match outcomes from Java example pair and official names.txt rows.
func TestPeopleRule_Rule(t *testing.T) {
	// Java: new PeopleRule(messages, language)
	rule := NewPeopleRule(nil)

	// correct sentences (Irish forms / non-keys) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Bhí sí cosúil le claíomh Dámaicléas ar crochadh sa spéir."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Abrahám anseo."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jornada gan eochair ainm."))))

	// Java example: Damocles → Dámaicléas
	checkPeople(t, rule, "Bhí sí cosúil le claíomh Damocles ar crochadh sa spéir.", "Dámaicléas")
	// shorter surface also used in ga_replace_rules_test
	checkPeople(t, rule, "claíomh Damocles ar crochadh", "Dámaicléas")

	// Official names.txt rows
	checkPeople(t, rule, "Abraham anseo.", "Abrahám")
	checkPeople(t, rule, "Adrian anseo.", "Aidrian")
	checkPeople(t, rule, "Alexander anseo.", "Alastar")
	checkPeople(t, rule, "Cicero anseo.", "Cicearó")

	// multi-suggestion key: Achilles=Aichill|Aicilléas
	matchesMulti := rule.Match(languagetool.AnalyzePlain("Achilles anseo."))
	require.Len(t, matchesMulti, 1)
	require.Equal(t, []string{"Aichill", "Aicilléas"},
		matchesMulti[0].GetSuggestedReplacements())
	require.Equal(t,
		"Ainm Béarla. \"Aichill, Aicilléas\".",
		matchesMulti[0].GetMessage())
	require.Equal(t, "Ainm Béarla", matchesMulti[0].GetShortMessage())

	// Java-visible createRuleMatch message/short for example-pair hit
	// (message uses replacements only; tokenStr unused)
	matches := rule.Match(languagetool.AnalyzePlain("Bhí sí cosúil le claíomh Damocles ar crochadh sa spéir."))
	require.Len(t, matches, 1)
	require.Equal(t, "Ainm Béarla", matches[0].GetShortMessage())
	require.Equal(t, "Ainm Béarla. \"Dámaicléas\".", matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, PeopleRuleID, base.GetID())
}

// checkPeople asserts one match with the given first replacement and exact message.
func checkPeople(t *testing.T, rule *PeopleRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	// Java: "Ainm Béarla. \"" + join(replacements) + "\"."
	// Message is built before uppercaseFirstChar on suggestions; single-rep keys use word as join.
	require.Equal(t,
		"Ainm Béarla. \""+word+"\".",
		matches[0].GetMessage(),
		"message for %q", sentence)
	require.Equal(t, "Ainm Béarla", matches[0].GetShortMessage(), "short for %q", sentence)
	require.Equal(t, PeopleRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of PeopleRule constructor + overrides.
func TestPeopleRule_Metadata(t *testing.T) {
	rule := NewPeopleRule(nil)

	require.Equal(t, PeopleRuleID, rule.GetID())
	require.Equal(t, "GA_PEOPLE", rule.GetID())
	require.Equal(t, "Ainm Béarla, m.sh., 'Damocles' in áit 'Dámaicléas'.", rule.GetDescription())
	require.Equal(t, "Ainm Béarla", rule.GetShort())
	// MessageFn with empty replacements (static surface for metadata probes)
	require.Equal(t, "Ainm Béarla. \"\".", rule.GetMessage())

	// Java getLocale() → Locale("GA")
	require.Equal(t, "GA", rule.GetLocaleCode())
	require.Equal(t, "GA", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.False(t, rule.SubRuleSpecificIDs)
	require.False(t, rule.IgnoreTaggedWords) // Java does not call setIgnoreTaggedWords
	require.Equal(t, "", rule.GetURL())

	// Java: Categories.TYPOS + ITSIssueType.Misspelling
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as /ga/names.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"Dámaicléas"}, ww["Damocles"])
	require.Equal(t, []string{"Aichill", "Aicilléas"}, ww["Achilles"])
	require.Equal(t, []string{"Abrahám"}, ww["Abraham"])
	require.Equal(t, []string{"Aidrian"}, ww["Adrian"])
	require.Equal(t, []string{"Alastar"}, ww["Alexander"])
	require.Equal(t, PeopleFilePath, "/ga/names.txt")
	require.Greater(t, len(ww), 50, "names.txt has many keys")

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_typo": "Botúin litrithe",
	}
	ruleMsgs := NewPeopleRule(gaMsgs)
	require.Equal(t, "TYPOS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Botúin litrithe", ruleMsgs.GetCategory().GetName())
	// Hardcoded GA description/short do not depend on messages
	require.Equal(t, "Ainm Béarla, m.sh., 'Damocles' in áit 'Dámaicléas'.", ruleMsgs.GetDescription())
	require.Equal(t, "Ainm Béarla", ruleMsgs.GetShort())

	// Java addExamplePair(Damocles → Dámaicléas)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Bhí sí cosúil le claíomh <marker>Damocles</marker> ar crochadh sa spéir.", inc[0].GetExample())
	require.Equal(t, []string{"Dámaicléas"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "Bhí sí cosúil le claíomh <marker>Dámaicléas</marker> ar crochadh sa spéir.", corr[0].GetExample())
}

// Load does not panic; resource parses.
func TestPeopleRule_Loads(t *testing.T) {
	rule := NewPeopleRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Greater(t, len(rule.GetWrongWords()), 0)
}

func TestPeopleRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := namesFS.ReadFile("data/names.txt")
	require.NoError(t, err)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	// this file: .../internal/languagetool/org/languagetool/rules/ga/*.go
	// seven Dir() steps reach repo root (ga→rules→languagetool→org→languagetool→internal→repo)
	repoRoot := thisFile
	for i := 0; i < 7; i++ {
		repoRoot = filepath.Dir(repoRoot)
	}
	official := filepath.Join(repoRoot,
		"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/names.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/names.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official ga/names.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded ga/data/names.txt must be byte-identical to official Java resource")

	m, err := rules.LoadSimpleReplaceWords(bytes.NewReader(embedded))
	require.NoError(t, err)
	require.Equal(t, []string{"Dámaicléas"}, m["Damocles"])
	require.Equal(t, []string{"Aichill", "Aicilléas"}, m["Achilles"])
	require.Equal(t, []string{"Abrahám"}, m["Abraham"])
}
