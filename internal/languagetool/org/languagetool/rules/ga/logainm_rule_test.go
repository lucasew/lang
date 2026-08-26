package ga

// Outcome twin of org.languagetool.rules.ga.LogainmRule
// (no dedicated Java *RuleTest class; bar matches IrishReplace / Prestandard /
// IrishFGBEq ACCEPT: full metadata + Match outcomes + exact getMessage text).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; CheckLemmas is false so lemmas are not consulted.
// Official /ga/placenames.txt keys are English placename surfaces (e.g. Dublin).
// Java setIgnoreTaggedWords() is true — tagged non-sentence-start tokens skip.

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

// Match outcomes from official placenames.txt rows (Java has no example pair).
func TestLogainmRule_Rule(t *testing.T) {
	// Java: new LogainmRule(messages, language)
	rule := NewLogainmRule(nil)

	// correct sentences (Irish forms / non-keys) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Baile Átha Cliath."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Mainistir na Féile."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jornada gan eochair logainm."))))

	// Official placenames.txt rows (Title Case keys as in file; surface match)
	checkLogainm(t, rule, "I live in Dublin.", "Dublin", "Baile Átha Cliath")
	checkLogainm(t, rule, "Abbeyfeale is a town.", "Abbeyfeale", "Mainistir na Féile")
	checkLogainm(t, rule, "Abbey is nearby.", "Abbey", "An Mhainistir")
	checkLogainm(t, rule, "Abkhaz region.", "Abkhaz", "An Abcáis")
	// Abkhaz|Abkhazia same replacement list
	checkLogainm(t, rule, "Abkhazia region.", "Abkhazia", "An Abcáis")

	// multi-suggestion key: Adamstown=Baile Adaim|Baile Ádaim|Baile Adáim|Maigh Arnaí
	matchesMulti := rule.Match(languagetool.AnalyzePlain("Adamstown area."))
	require.Len(t, matchesMulti, 1)
	require.Equal(t, []string{"Baile Adaim", "Baile Ádaim", "Baile Adáim", "Maigh Arnaí"},
		matchesMulti[0].GetSuggestedReplacements())
	require.Equal(t,
		"Is logainm Béarla é \"Adamstown\" ar \"Baile Adaim, Baile Ádaim, Baile Adáim, Maigh Arnaí\".",
		matchesMulti[0].GetMessage())
	require.Equal(t, "Logainm Béarla", matchesMulti[0].GetShortMessage())

	// Java-visible createRuleMatch message/short for Dublin hit
	// (message uses both tokenStr and replacements)
	matches := rule.Match(languagetool.AnalyzePlain("I live in Dublin."))
	require.Len(t, matches, 1)
	require.Equal(t, "Logainm Béarla", matches[0].GetShortMessage())
	require.Equal(t,
		"Is logainm Béarla é \"Dublin\" ar \"Baile Átha Cliath\".",
		matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, LogainmRuleID, base.GetID())
}

// checkLogainm asserts one match with the given first replacement and exact message.
// tokenStr is the surface form Java passes to getMessage (before title-case of suggestions).
func checkLogainm(t *testing.T, rule *LogainmRule, sentence, tokenStr, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	// Java: "Is logainm Béarla é \"" + tokenStr + "\" ar \"" + join(replacements) + "\"."
	// Message is built before uppercaseFirstChar on suggestions; single-rep keys use word as join.
	require.Equal(t,
		"Is logainm Béarla é \""+tokenStr+"\" ar \""+word+"\".",
		matches[0].GetMessage(),
		"message for %q", sentence)
	require.Equal(t, "Logainm Béarla", matches[0].GetShortMessage(), "short for %q", sentence)
	require.Equal(t, LogainmRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of LogainmRule constructor + overrides.
func TestLogainmRule_Metadata(t *testing.T) {
	rule := NewLogainmRule(nil)

	require.Equal(t, LogainmRuleID, rule.GetID())
	require.Equal(t, "GA_LOGAINM", rule.GetID())
	require.Equal(t, "Logainm Béarla, m.sh., 'Dublin' in áit 'Baile Átha Cliath'.", rule.GetDescription())
	require.Equal(t, "Logainm Béarla", rule.GetShort())
	// MessageFn with empty token + empty replacements (static surface for metadata probes)
	require.Equal(t, "Is logainm Béarla é \"\" ar \"\".", rule.GetMessage())

	// Java getLocale() → Locale("GA")
	require.Equal(t, "GA", rule.GetLocaleCode())
	require.Equal(t, "GA", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.True(t, rule.IgnoreTaggedWords) // unlike IrishFGBEq / IrishReplace
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, LogainmURL, rule.GetURL())
	require.Equal(t, "https://www.logainm.ie/ga/", rule.GetURL())

	// Java: Categories.TYPOS + ITSIssueType.Misspelling
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as /ga/placenames.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"Baile Átha Cliath"}, ww["Dublin"])
	require.Equal(t, []string{"An Mhainistir"}, ww["Abbey"])
	require.Equal(t, []string{"Mainistir na Féile"}, ww["Abbeyfeale"])
	require.Equal(t, []string{"An Abcáis"}, ww["Abkhaz"])
	require.Equal(t, []string{"An Abcáis"}, ww["Abkhazia"])
	require.Equal(t, []string{"Baile Adaim", "Baile Ádaim", "Baile Adáim", "Maigh Arnaí"}, ww["Adamstown"])
	require.Equal(t, LogainmFilePath, "/ga/placenames.txt")
	require.Greater(t, len(ww), 100, "placenames.txt has many keys")

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_typo": "Botúin litrithe",
	}
	ruleMsgs := NewLogainmRule(gaMsgs)
	require.Equal(t, "TYPOS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Botúin litrithe", ruleMsgs.GetCategory().GetName())
	// Hardcoded GA description/short do not depend on messages
	require.Equal(t, "Logainm Béarla, m.sh., 'Dublin' in áit 'Baile Átha Cliath'.", ruleMsgs.GetDescription())
	require.Equal(t, "Logainm Béarla", ruleMsgs.GetShort())

	// Java has no addExamplePair
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
}

// Java setIgnoreTaggedWords(): tagged non-sentence-start surfaces that would
// match WrongWords are skipped. Mirrors abstract_simple_replace_ignore_test.go.
func TestLogainmRule_IgnoreTaggedWords(t *testing.T) {
	rule := NewLogainmRule(nil)
	require.True(t, rule.IgnoreTaggedWords)

	// Untagged "Dublin" → match
	untagged := rule.Match(languagetool.AnalyzePlain("Dublin."))
	require.Len(t, untagged, 1)
	require.Equal(t, "Baile Átha Cliath", untagged[0].GetSuggestedReplacements()[0])

	// Hand-built sentence: SENT_START + tagged "Dublin" → skip
	ss := languagetool.SentenceStartTagName
	startTok := languagetool.NewAnalyzedToken("", &ss, nil)
	startR := languagetool.NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)

	pos := "NN"
	dublinTok := languagetool.NewAnalyzedToken("Dublin", &pos, nil)
	dublinR := languagetool.NewAnalyzedTokenReadingsAt(dublinTok, 0)
	require.True(t, dublinR.IsTagged(), "token with POS must be tagged")

	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{startR, dublinR})
	require.Empty(t, rule.Match(sent), "tagged non-sentence-start placename must be skipped")

	// Same surface untagged (hand-built) still matches
	plainDublin := languagetool.NewAnalyzedToken("Dublin", nil, nil)
	plainR := languagetool.NewAnalyzedTokenReadingsAt(plainDublin, 0)
	require.False(t, plainR.IsTagged())
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{startR, plainR})
	ms := rule.Match(sent2)
	require.Len(t, ms, 1)
	require.Equal(t, "Baile Átha Cliath", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t,
		"Is logainm Béarla é \"Dublin\" ar \"Baile Átha Cliath\".",
		ms[0].GetMessage())
}

// Load does not panic; resource parses.
func TestLogainmRule_Loads(t *testing.T) {
	rule := NewLogainmRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Greater(t, len(rule.GetWrongWords()), 0)
}

func TestLogainmRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := placenamesFS.ReadFile("data/placenames.txt")
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
		"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/placenames.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/placenames.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official ga/placenames.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded ga/data/placenames.txt must be byte-identical to official Java resource")

	m, err := rules.LoadSimpleReplaceWords(bytes.NewReader(embedded))
	require.NoError(t, err)
	require.Equal(t, []string{"Baile Átha Cliath"}, m["Dublin"])
	require.Equal(t, []string{"An Abcáis"}, m["Abkhaz"])
}
