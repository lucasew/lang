package ga

// Outcome twin of org.languagetool.rules.ga.IrishSpecificCaseRule
// (no dedicated Java *RuleTest class; bar matches RussianSpecificCaseRule ACCEPT:
// full metadata + Match outcomes + exact message/short + resource).
//
// Analysis: Java uses JLanguageTool(ga).getAnalyzedSentence. For this surface
// casing rule, AnalyzePlain token boundaries match Java for the test strings
// (same approach as accepted English/Russian SpecificCaseRule twins).

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

const (
	msgInitialCapital = "Más ainmfhocal dílis é, scríobh é i gceannlitreacha."
	msgOtherCapital   = "If the term is a proper noun, use the suggested capitalization."
)

func TestIrishSpecificCaseRule_Rule(t *testing.T) {
	// Java: new IrishSpecificCaseRule(TestTools.getEnglishMessages())
	rule := NewIrishSpecificCaseRule(nil)

	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "good %q", s)
	}
	assertBad := func(s string) []*rules.RuleMatch {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 1, len(matches), "bad %q", s)
		return matches
	}

	// Correct proper casing → 0 matches
	assertGood("Béal Feirste")
	assertGood("mBéal Feirste")
	assertGood("Sinn Féin")
	assertGood("Fianna Fáil")
	assertGood("Bealach na Bó Finne")

	// Java example: mbéal Feirste → mBéal Feirste
	// proper "mBéal Feirste": Utils treats both words as starting uppercase → Irish msg
	matchesEx := assertBad("i mbéal Feirste é.")
	require.Equal(t, []string{"mBéal Feirste"}, matchesEx[0].GetSuggestedReplacements())
	require.Equal(t, msgInitialCapital, matchesEx[0].GetMessage())
	// Java: new RuleMatch(..., msg) — shortMessage null
	require.Equal(t, "", matchesEx[0].GetShortMessage())

	// Full example-pair sentence
	matchesFull := assertBad("Rugadh agus tógadh i mbéal Feirste é.")
	require.Equal(t, []string{"mBéal Feirste"}, matchesFull[0].GetSuggestedReplacements())
	require.Equal(t, msgInitialCapital, matchesFull[0].GetMessage())
	require.Equal(t, "", matchesFull[0].GetShortMessage())

	// All-words-uppercase proper spelling → Irish message
	matchesSF := assertBad("sinn féin anseo.")
	require.Equal(t, []string{"Sinn Féin"}, matchesSF[0].GetSuggestedReplacements())
	require.Equal(t, msgInitialCapital, matchesSF[0].GetMessage())
	require.Equal(t, "", matchesSF[0].GetShortMessage())

	matchesFF := assertBad("fianna fáil anseo.")
	require.Equal(t, []string{"Fianna Fáil"}, matchesFF[0].GetSuggestedReplacements())
	require.Equal(t, msgInitialCapital, matchesFF[0].GetMessage())

	// Mixed proper casing ("na" lowercase) → English message branch
	matchesMixed := assertBad("bealach na bó finne anseo.")
	require.Equal(t, []string{"Bealach na Bó Finne"}, matchesMixed[0].GetSuggestedReplacements())
	require.Equal(t, msgOtherCapital, matchesMixed[0].GetMessage())
	require.Equal(t, "", matchesMixed[0].GetShortMessage())

	// béal Feirste → Béal Feirste (all words upper in proper → Irish msg)
	matchesBF := assertBad("i béal Feirste é.")
	require.Equal(t, []string{"Béal Feirste"}, matchesBF[0].GetSuggestedReplacements())
	require.Equal(t, msgInitialCapital, matchesBF[0].GetMessage())

	// All-uppercase input is skipped (Utils.isAllUppercase)
	assertGood("BÉAL FEIRSTE")
}

// Java IrishSpecificCaseRule ctor metadata: CASING, Misspelling, example pair,
// resource /ga/specific_case.txt, empty short (5-arg RuleMatch).
func TestIrishSpecificCaseRule_Metadata(t *testing.T) {
	rule := NewIrishSpecificCaseRule(nil)
	require.Equal(t, IrishSpecificCaseRuleID, rule.GetID())
	require.Equal(t, "GA_SPECIFIC_CASE", rule.GetID())
	require.Equal(t, IrishSpecificCaseFilePath, "/ga/specific_case.txt")
	require.Equal(t, "Checks upper/lower case spelling of some proper nouns", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "CASING", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	// Java does not setUrl
	require.Equal(t, "", rule.GetURL())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Rugadh agus tógadh i <marker>mbéal Feirste</marker> é.", inc[0].GetExample())
	require.Equal(t, []string{"mBéal Feirste"}, inc[0].GetCorrections())
	require.Equal(t, "Rugadh agus tógadh i <marker>mBéal Feirste</marker> é.", rule.GetCorrectExamples()[0].GetExample())

	// Resource twin of Java getFromResourceDirAsLines("/ga/specific_case.txt")
	m, maxLen := loadSpecificCase()
	require.NotEmpty(t, m)
	require.GreaterOrEqual(t, maxLen, 2)
	// known phrases from official specific_case.txt
	require.Equal(t, "Béal Feirste", m["béal feirste"])
	require.Equal(t, "mBéal Feirste", m["mbéal feirste"])
	require.Equal(t, "Sinn Féin", m["sinn féin"])
	require.Equal(t, "Fianna Fáil", m["fianna fáil"])
	require.Equal(t, "Bealach na Bó Finne", m["bealach na bó finne"])

	// Java message branches (exact strings); short empty (no getShortMessage override)
	require.Equal(t, msgInitialCapital, rule.InitialCapitalMessage)
	require.Equal(t, msgOtherCapital, rule.OtherCapitalizationMessage)
	require.Equal(t, "", rule.ShortMsg)
}

// Load does not panic; resource parses.
func TestIrishSpecificCaseRule_Loads(t *testing.T) {
	rule := NewIrishSpecificCaseRule(nil)
	require.NotNil(t, rule)
	m, maxLen := loadSpecificCase()
	require.NotEmpty(t, m)
	require.GreaterOrEqual(t, maxLen, 1)
}

func TestIrishSpecificCaseRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := specificCaseFS.ReadFile("data/specific_case.txt")
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
		"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/resource/ga/specific_case.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/resource/ga/specific_case.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official ga/specific_case.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded ga/data/specific_case.txt must be byte-identical to official Java resource")

	m, maxLen, err := rules.LoadSpecificCasePhrases(bytes.NewReader(embedded))
	require.NoError(t, err)
	require.Equal(t, "Béal Feirste", m["béal feirste"])
	require.Equal(t, "mBéal Feirste", m["mbéal feirste"])
	require.Equal(t, "Sinn Féin", m["sinn féin"])
	require.Equal(t, "Fianna Fáil", m["fianna fáil"])
	require.GreaterOrEqual(t, maxLen, 4) // "Bealach na Bó Finne" = 4 tokens
	require.Equal(t, 44, len(m))         // 48 lines file - comments - empty = phrase count
}
