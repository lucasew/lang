package ga

// Outcome twin of org.languagetool.rules.ga.SpacesRule
// (no dedicated Java *RuleTest class; bar matches People / Logainm / IrishFGBEq ACCEPT:
// full metadata + Match outcomes + exact getMessage text).
//
// Analysis: AnalyzePlain (WordTokenizer + SENT_START) is equivalent for this
// rule’s surface/dictionary path. AbstractSimpleReplaceRule matches token
// surfaces against wrongWords; CheckLemmas is false so lemmas are not consulted.
// Official /ga/spaces.txt keys are missing-space surfaces (e.g. abhfuil, agan).
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

// Match outcomes from official spaces.txt rows (Java has no example pair).
func TestSpacesRule_Rule(t *testing.T) {
	// Java: new SpacesRule(messages, language)
	rule := NewSpacesRule(nil)

	// correct sentences (fixed forms / non-keys) → 0 matches
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("rud a bhfuil ann"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("a bhaineann leis"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("ag an doras"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Jornada gan eochair spaces."))))

	// Official spaces.txt rows
	checkSpaces(t, rule, "rud abhfuil ann", "a bhfuil")
	checkSpaces(t, rule, "sé abhaineann leis", "a bhaineann")
	checkSpaces(t, rule, "agan doras", "ag an")
	checkSpaces(t, rule, "abheith ann", "a bheith")
	checkSpaces(t, rule, "achur i gcrích", "a chur")
	checkSpaces(t, rule, "aonduine eile", "aon duine")
	checkSpaces(t, rule, "araghaidh an", "ar aghaidh")
	checkSpaces(t, rule, "chomhmaith leis", "chomh maith")

	// Java-visible createRuleMatch message/short for abhfuil hit
	// (message uses replacements only; tokenStr unused)
	matches := rule.Match(languagetool.AnalyzePlain("rud abhfuil ann"))
	require.Len(t, matches, 1)
	require.Equal(t, "Spás ar iarraidh", matches[0].GetShortMessage())
	require.Equal(t, "Spás ar iarraidh: \"a bhfuil\".", matches[0].GetMessage())
	// no useSubRuleSpecificIds — match rule is base AbstractSimpleReplaceRule
	_, isSpecific := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSpecific, "sub-rule specific id must not be used")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule)
	require.True(t, ok, "base AbstractSimpleReplaceRule expected without sub-rule ids")
	require.Equal(t, SpacesRuleID, base.GetID())
}

// checkSpaces asserts one match with the given first replacement and exact message.
func checkSpaces(t *testing.T, rule *SpacesRule, sentence, word string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 1, len(matches), "Invalid matches.length while checking sentence: %s", sentence)
	require.GreaterOrEqual(t, len(matches[0].GetSuggestedReplacements()), 1,
		"Invalid replacement count while checking sentence: %s", sentence)
	require.Equal(t, word, matches[0].GetSuggestedReplacements()[0],
		"Invalid suggested replacement while checking sentence: %s", sentence)
	// Java: "Spás ar iarraidh: \"" + join(replacements) + "\"."
	// Message is built before uppercaseFirstChar on suggestions; single-rep keys use word as join.
	require.Equal(t,
		"Spás ar iarraidh: \""+word+"\".",
		matches[0].GetMessage(),
		"message for %q", sentence)
	require.Equal(t, "Spás ar iarraidh", matches[0].GetShortMessage(), "short for %q", sentence)
	require.Equal(t, SpacesRuleID, matches[0].GetRule().(interface{ GetID() string }).GetID())
}

// Field twin of SpacesRule constructor + overrides.
func TestSpacesRule_Metadata(t *testing.T) {
	rule := NewSpacesRule(nil)

	require.Equal(t, SpacesRuleID, rule.GetID())
	require.Equal(t, "GA_SPASANNA", rule.GetID())
	require.Equal(t, "Spás ar iarraidh", rule.GetDescription())
	require.Equal(t, "Spás ar iarraidh", rule.GetShort())
	// MessageFn with empty replacements (static surface for metadata probes)
	require.Equal(t, "Spás ar iarraidh: \"\".", rule.GetMessage())

	// Java getLocale() → Locale("GA")
	require.Equal(t, "GA", rule.GetLocaleCode())
	require.Equal(t, "GA", rule.LanguageCode)
	require.False(t, rule.CaseSensitive)
	require.False(t, rule.CheckLemmas)
	require.True(t, rule.IgnoreTaggedWords) // Java setIgnoreTaggedWords()
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, "", rule.GetURL())

	// Java: Categories.TYPOS + ITSIssueType.Misspelling
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOS", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Official dictionary loaded (same keys as /ga/spaces.txt)
	ww := rule.GetWrongWords()
	require.NotNil(t, ww)
	require.Equal(t, []string{"a bhfuil"}, ww["abhfuil"])
	require.Equal(t, []string{"a bhaineann"}, ww["abhaineann"])
	require.Equal(t, []string{"ag an"}, ww["agan"])
	require.Equal(t, []string{"a bheith"}, ww["abheith"])
	require.Equal(t, []string{"a chur"}, ww["achur"])
	require.Equal(t, []string{"aon duine"}, ww["aonduine"])
	require.Equal(t, []string{"ar aghaidh"}, ww["araghaidh"])
	require.Equal(t, []string{"chomh maith"}, ww["chomhmaith"])
	require.Equal(t, SpacesFilePath, "/ga/spaces.txt")
	require.Equal(t, 87, len(ww), "all spaces.txt keys must load")

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_typo": "Botúin litrithe",
	}
	ruleMsgs := NewSpacesRule(gaMsgs)
	require.Equal(t, "TYPOS", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Botúin litrithe", ruleMsgs.GetCategory().GetName())
	// Hardcoded GA description/short do not depend on messages
	require.Equal(t, "Spás ar iarraidh", ruleMsgs.GetDescription())
	require.Equal(t, "Spás ar iarraidh", ruleMsgs.GetShort())

	// Java has no addExamplePair
	require.Empty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.GetCorrectExamples())
}

// Java setIgnoreTaggedWords(): tagged non-sentence-start surfaces that would
// match WrongWords are skipped. Mirrors LogainmRule_IgnoreTaggedWords.
func TestSpacesRule_IgnoreTaggedWords(t *testing.T) {
	rule := NewSpacesRule(nil)
	require.True(t, rule.IgnoreTaggedWords)

	// Untagged "abhfuil" → match
	untagged := rule.Match(languagetool.AnalyzePlain("abhfuil."))
	require.Len(t, untagged, 1)
	require.Equal(t, "a bhfuil", untagged[0].GetSuggestedReplacements()[0])

	// Hand-built sentence: SENT_START + tagged "abhfuil" → skip
	ss := languagetool.SentenceStartTagName
	startTok := languagetool.NewAnalyzedToken("", &ss, nil)
	startR := languagetool.NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)

	pos := "Verb:VTI:PresInd:Aut"
	spaceTok := languagetool.NewAnalyzedToken("abhfuil", &pos, nil)
	spaceR := languagetool.NewAnalyzedTokenReadingsAt(spaceTok, 0)
	require.True(t, spaceR.IsTagged(), "token with POS must be tagged")

	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{startR, spaceR})
	require.Empty(t, rule.Match(sent), "tagged non-sentence-start spaces key must be skipped")

	// Same surface untagged (hand-built) still matches
	plainTok := languagetool.NewAnalyzedToken("abhfuil", nil, nil)
	plainR := languagetool.NewAnalyzedTokenReadingsAt(plainTok, 0)
	require.False(t, plainR.IsTagged())
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{startR, plainR})
	ms := rule.Match(sent2)
	require.Len(t, ms, 1)
	require.Equal(t, "a bhfuil", ms[0].GetSuggestedReplacements()[0])
	require.Equal(t,
		"Spás ar iarraidh: \"a bhfuil\".",
		ms[0].GetMessage())
}

// Load does not panic; resource parses.
func TestSpacesRule_Loads(t *testing.T) {
	rule := NewSpacesRule(nil)
	require.NotNil(t, rule)
	require.NotNil(t, rule.GetWrongWords())
	require.Greater(t, len(rule.GetWrongWords()), 0)
}

func TestSpacesRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := spacesFS.ReadFile("data/spaces.txt")
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
		"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/spaces.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/spaces.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official ga/spaces.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded ga/data/spaces.txt must be byte-identical to official Java resource")

	m, err := rules.LoadSimpleReplaceWords(bytes.NewReader(embedded))
	require.NoError(t, err)
	require.Equal(t, []string{"a bhfuil"}, m["abhfuil"])
	require.Equal(t, []string{"ag an"}, m["agan"])
	require.Equal(t, 87, len(m))
}
