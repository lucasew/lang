package ga

// Twin of languagetool-language-modules/ga/src/test/java/org/languagetool/rules/ga/DhaNoBeirtRuleTest.java
// plus metadata / resource / control-flow outcome twins for validator-ready fidelity.

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Full twin of DhaNoBeirtRuleTest.testRule (Java counts via assertCorrect / assertIncorrect).
func TestDhaNoBeirtRule_Rule(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)

	// assertCorrect
	assertCorrect(t, rule, "Seo abairt bheag.")
	assertCorrect(t, rule, "Tá beirt dheartháireacha agam.")

	// assertIncorrect — Java tagger: dheartháireacha lemma deartháir ∈ people.txt
	// (inject lemma; no de-lenition invent in rule).
	assertIncorrect(t, rule, withLemma("Tá dhá dheartháireacha agam.", "dheartháireacha", "deartháir"), 1)
	// "ab" is listed in people.txt as surface → dhá + déag dual matches.
	assertIncorrect(t, rule, languagetool.AnalyzePlain("Seo dhá ab déag"), 2)
	// list of words between dhá and déag
	assertIncorrect(t, rule,
		withLemma("Tá dhá dheartháireacha níos aosta déag agam.", "dheartháireacha", "deartháir"), 2)
}

func assertCorrect(t *testing.T, rule *DhaNoBeirtRule, sentence string) {
	t.Helper()
	matches := rule.Match(languagetool.AnalyzePlain(sentence))
	require.Equal(t, 0, len(matches), "expected 0 matches for %q", sentence)
}

func assertIncorrect(t *testing.T, rule *DhaNoBeirtRule, sentence *languagetool.AnalyzedSentence, expected int) {
	t.Helper()
	matches := rule.Match(sentence)
	require.Equal(t, expected, len(matches), "expected %d matches for %q", expected, sentence.GetText())
}

// Fail-closed without lemma: lenited plural not in people.txt as surface.
// Java isPerson does not invent de-lenition; same here.
func TestDhaNoBeirtRule_FailClosedWithoutLemma(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tá dhá dheartháireacha agam."))))
}

// Message / short-message twins for simple replace and déag dual path.
func TestDhaNoBeirtRule_Messages(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)

	// Simple: dhá + person (lemma) → one match, personal-number message.
	sent := withLemma("Tá dhá dheartháireacha agam.", "dheartháireacha", "deartháir")
	ms := rule.Match(sent)
	require.Len(t, ms, 1)
	require.Equal(t, "Uimhir phearsanta", ms[0].GetShortMessage())
	require.Equal(t,
		"Ba chóir duit <suggestion>beirt dheartháireacha</suggestion> a scríobh",
		ms[0].GetMessage())
	// Span is the number token "dhá" (prevTokenIndex+1 under normal loop)
	var dhaTok *languagetool.AnalyzedTokenReadings
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok != nil && tok.GetToken() == "dhá" {
			dhaTok = tok
			break
		}
	}
	require.NotNil(t, dhaTok)
	require.Equal(t, dhaTok.GetStartPos(), ms[0].GetFromPos())
	require.Equal(t, dhaTok.GetEndPos(), ms[0].GetToPos())

	// Dual: dhá + ab + déag → two matches (dháréag + delete déag)
	ms2 := rule.Match(languagetool.AnalyzePlain("Seo dhá ab déag"))
	require.Len(t, ms2, 2)
	require.Equal(t, "Uimhir phearsanta", ms2[0].GetShortMessage())
	require.Equal(t, "Uimhir phearsanta", ms2[1].GetShortMessage())
	require.Equal(t,
		"Ba chóir duit <suggestion>dháréag</suggestion> a scríobh",
		ms2[0].GetMessage())
	require.Equal(t,
		"Ba chóir duit \"déag\" a scriosadh.",
		ms2[1].GetMessage())
}

// Length-2 gate: number+person at end of sentence (no token after person) → no match.
// Java: i < tokens.length - 2 required (person must not be last non-whitespace token).
func TestDhaNoBeirtRule_Length2Gate(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)
	// Tokens without whitespace: SENT_START, Tá, dhá, ab — person is last real token
	// (no trailing punctuation token). i for dhá fails i < len-2.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tá dhá ab"))))
	// With trailing period: person is not last → match fires (period is token after person).
	ms := rule.Match(languagetool.AnalyzePlain("Tá dhá ab."))
	require.Len(t, ms, 1)
	require.Equal(t,
		"Ba chóir duit <suggestion>beirt ab</suggestion> a scríobh",
		ms[0].GetMessage())
}

// Exact-case map get: isNumber is case-insensitive, but map.get uses exact token case.
// Uppercased "Dhá" → isNumber true, get miss → Java builds message with literal "null".
func TestDhaNoBeirtRule_ExactCaseMapGet(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)
	ms := rule.Match(languagetool.AnalyzePlain("Tá Dhá ab."))
	require.Len(t, ms, 1)
	require.Equal(t,
		"Ba chóir duit <suggestion>null ab</suggestion> a scríobh",
		ms[0].GetMessage())
}

// Sticky replacement: after dhá…déag sets replacement="dháréag", a later number+person
// does not add a simple-replace match (Java never clears replacement).
func TestDhaNoBeirtRule_StickyReplacement(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)
	// First: dhá ab déag → 2 matches; sticky replacement = dháréag
	// Later: trí ab . would be a simple match if replacement were cleared — must not fire.
	// Sentence needs person after each number and a token after person for the gate.
	ms := rule.Match(languagetool.AnalyzePlain("Seo dhá ab déag agus trí ab ann."))
	// Only the two déag-path matches; no third simple-replace on "trí"
	require.Len(t, ms, 2)
	require.Contains(t, ms[0].GetMessage(), "dháréag")
	require.Contains(t, ms[1].GetMessage(), "déag")
}

// Lemma branch is exact (no toLower invent): uppercase lemma not in set → fail closed.
func TestDhaNoBeirtRule_LemmaExactNoToLower(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)
	// people.txt has lowercase "deartháir" after load; lemma "Deartháir" must not match.
	require.Equal(t, 0, len(rule.Match(
		withLemma("Tá dhá dheartháireacha agam.", "dheartháireacha", "Deartháir"))))
	// exact lowercase lemma hits
	require.Equal(t, 1, len(rule.Match(
		withLemma("Tá dhá dheartháireacha agam.", "dheartháireacha", "deartháir"))))
}

// Field twin of constructor + overrides (id, description, category, ITS, example).
func TestDhaNoBeirtRule_Metadata(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)

	require.Equal(t, DhaNoBeirtRuleID, rule.GetID())
	require.Equal(t, "GA_DHA_NO_BEIRT", rule.GetID())
	require.Equal(t, "'dhá' nó 'beirt'", rule.GetDescription())

	// Java: Categories.MISC + ITSIssueType.Misspelling
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	// Category name honors supplied ResourceBundle map
	gaMsgs := map[string]string{
		"category_misc": "Ilghnéitheach",
	}
	ruleMsgs := NewDhaNoBeirtRule(gaMsgs)
	require.Equal(t, "MISC", ruleMsgs.GetCategory().GetID().String())
	require.Equal(t, "Ilghnéitheach", ruleMsgs.GetCategory().GetName())
	// Hardcoded description does not depend on messages
	require.Equal(t, "'dhá' nó 'beirt'", ruleMsgs.GetDescription())

	// Java addExamplePair(dhá → beirt)
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Tá <marker>dhá</marker> dheartháireacha agam.", inc[0].GetExample())
	require.Equal(t, []string{"beirt"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t, "Tá <marker>beirt</marker> dheartháireacha agam.", corr[0].GetExample())
}

// DhaNoBeirtData static twin: 9 number entries + people set loaded.
func TestDhaNoBeirtData_NumberMapAndDaoine(t *testing.T) {
	var d DhaNoBeirtData
	nums := d.GetNumberReplacements()
	require.Equal(t, 9, len(nums))
	require.Equal(t, "beirt", nums["dhá"])
	require.Equal(t, "triúr", nums["trí"])
	require.Equal(t, "ceathrar", nums["ceathair"])
	require.Equal(t, "cúigear", nums["cúig"])
	require.Equal(t, "seisear", nums["sé"])
	require.Equal(t, "seachtar", nums["seacht"])
	require.Equal(t, "ochtar", nums["ocht"])
	require.Equal(t, "naonúr", nums["naoi"])
	require.Equal(t, "deichniúr", nums["deich"])

	daoine := d.GetDaoine()
	require.NotNil(t, daoine)
	require.True(t, daoine["ab"])
	require.True(t, daoine["deartháir"])
	require.Equal(t, DhaNoBeirtPeoplePath, "/ga/people.txt")
	require.Greater(t, len(daoine), 50, "people.txt has many entries")
}

func TestDhaNoBeirtRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := peopleFS.ReadFile("data/people.txt")
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
		"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/people.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/rules/ga/people.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official ga/people.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded ga/data/people.txt must be byte-identical to official Java resource")
}

func withLemma(text, surface, lemma string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(text, func(tok string) []languagetool.TokenTag {
		if strings.EqualFold(tok, surface) {
			return []languagetool.TokenTag{{POS: "Noun", Lemma: lemma}}
		}
		return nil
	})
}
