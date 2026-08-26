package ga

// Outcome twin for org.languagetool.rules.ga.CompoundRule
// (no dedicated Java *CompoundRuleTest for GA — BR/RO-parity outcome matrix).
//
// Java ctor metadata:
//   id GA_COMPOUNDS
//   description "Focail fhleiscínithe, e.g., Moltar 'ró-úsáid' seachas 'ró úsáid'"
//   Categories.MISC, ITSIssueType.Misspelling (AbstractCompoundRule defaults)
//   4 Irish messages + shortMessage
//   example: mí úsáid → mí-úsáid
//   getCompoundRuleData → CompoundRuleData("/ga/compounds.txt")
//
// Official compounds.txt has only two '*' (dash-only) entries: ró-úsáid*, mí-úsáid*.
// Analysis: AnalyzePlain (WordTokenizer) — surface keys match Java for these phrases.

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
	withHyphenMsg    = "Litrítear an focal seo le fleiscín de ghnáth."
	withoutHyphenMsg = "Litrítear an focal seo mar fhocal amháin de ghnáth."
	withOrWithoutMsg = "Litrítear an nath seo mar fhocal amháin nó le fleiscín."
	shortMsg         = "Fadhb leis an bhfleiscíniú"
	gaCompoundDesc   = "Focail fhleiscínithe, e.g., Moltar 'ró-úsáid' seachas 'ró úsáid'"
)

func TestCompoundRule_Metadata(t *testing.T) {
	r := NewCompoundRule(nil)
	// Java getId / getDescription
	require.Equal(t, CompoundRuleID, r.GetID())
	require.Equal(t, "GA_COMPOUNDS", r.GetID())
	require.Equal(t, gaCompoundDesc, r.GetDescription())
	// Java AbstractCompoundRule ctor: Categories.MISC + ITSIssueType.Misspelling
	require.NotNil(t, r.GetCategory())
	require.Equal(t, rules.CatMisc.GetCategory(nil).GetID(), r.GetCategory().GetID())
	require.Equal(t, rules.ITSMisspelling, r.GetLocQualityIssueType())
	// Java 4-arg super shortMessage + GA messages
	require.Equal(t, shortMsg, r.ShortDesc)
	require.Equal(t, withHyphenMsg, r.WithHyphenMessage)
	require.Equal(t, withoutHyphenMsg, r.WithoutHyphenMessage)
	require.Equal(t, withOrWithoutMsg, r.WithOrWithoutHyphenMessage)
	// Java AbstractCompoundRule default / field true
	require.True(t, r.SentenceStartsWithUpperCase)
	// Java: no useSubRuleSpecificIds()
	require.False(t, r.SubRuleSpecificIDs)
	// Java: no getUrl override
	require.Equal(t, "", r.GetURL())
	// Java getCompoundRuleData singleton from /ga/compounds.txt
	require.Equal(t, CompoundFilePath, "/ga/compounds.txt")
	require.NotNil(t, r.GetCompoundRuleData())
	require.Same(t, loadCompounds(), r.GetCompoundRuleData())
	require.Same(t, r.Data, r.GetCompoundRuleData())
	// data contains official entries (hyphens → spaces in CompoundRuleData)
	require.True(t, r.GetCompoundRuleData().ContainsIncorrect("mí úsáid"))
	require.True(t, r.GetCompoundRuleData().ContainsDash("mí úsáid"))
	require.False(t, r.GetCompoundRuleData().ContainsJoined("mí úsáid"))
	require.True(t, r.GetCompoundRuleData().ContainsIncorrect("ró úsáid"))
	require.True(t, r.GetCompoundRuleData().ContainsDash("ró úsáid"))
	require.False(t, r.GetCompoundRuleData().ContainsJoined("ró úsáid"))
	// Java example pair
	inex := r.GetIncorrectExamples()
	require.Len(t, inex, 1)
	require.Equal(t, []string{"mí-úsáid"}, inex[0].GetCorrections())
	require.Equal(t, "Tá <marker>mí úsáid</marker> fhisiciúil i gceist.", inex[0].GetExample())
	cex := r.GetCorrectExamples()
	require.Len(t, cex, 1)
	require.Equal(t, "Tá <marker>mí-úsáid</marker> fhisiciúil i gceist.", cex[0].GetExample())
	// EstimateContextForSureMatch → 1 (AbstractCompoundRule)
	require.Equal(t, 1, r.EstimateContextForSureMatch())
}

func TestCompoundRule_Rule(t *testing.T) {
	// All official GA compounds.txt entries end with '*' → dash-only → WithHyphenMessage.
	rule := NewCompoundRule(nil)

	check := func(expectedErrors int, text string, expSuggestions ...string) {
		t.Helper()
		// Surface analysis sufficient for compounds (same as SK / BR AnalyzePlain paths).
		sent := languagetool.AnalyzePlain(text)
		require.NotNil(t, sent, "AnalyzePlain(%q)", text)
		matches := rule.Match(sent)
		require.Equal(t, expectedErrors, len(matches), "text %q got %v", text, matches)
		if len(expSuggestions) > 0 {
			require.Equal(t, 1, expectedErrors)
			require.Equal(t, expSuggestions, matches[0].GetSuggestedReplacements(), "text %q", text)
			require.Equal(t, withHyphenMsg, matches[0].GetMessage(), "text %q", text)
			require.Equal(t, shortMsg, matches[0].GetShortMessage(), "text %q", text)
			// no useSubRuleSpecificIds → specific id falls back to rule id
			require.Equal(t, "GA_COMPOUNDS", matches[0].GetSpecificRuleId())
		}
	}

	// --- Official example pair (Java addExamplePair) ---
	check(1, "Tá mí úsáid fhisiciúil i gceist.", "mí-úsáid")
	check(0, "Tá mí-úsáid fhisiciúil i gceist.")

	// --- Both compounds.txt entries ---
	// mí-úsáid*
	check(1, "mí úsáid", "mí-úsáid")
	check(0, "mí-úsáid")
	// sentence-start capitalisation (SentenceStartsWithUpperCase=true)
	check(1, "Mí úsáid", "Mí-úsáid")
	// ró-úsáid*
	check(1, "ró úsáid", "ró-úsáid")
	check(0, "ró-úsáid")
	check(1, "Ró úsáid", "Ró-úsáid")
	// embedded in a short sentence
	check(1, "Tá ró úsáid ann.", "ró-úsáid")
	check(0, "Tá ró-úsáid ann.")
}

// Load does not panic; resource parses both official entries.
func TestCompoundRule_Loads(t *testing.T) {
	rule := NewCompoundRule(nil)
	require.NotNil(t, rule)
	d := rule.GetCompoundRuleData()
	require.NotNil(t, d)
	require.True(t, d.ContainsIncorrect("mí úsáid"))
	require.True(t, d.ContainsIncorrect("ró úsáid"))
}

func TestCompoundRule_OfficialResourceByteIdentical(t *testing.T) {
	// Embedded data must be the official Java resource, byte-identical.
	embedded, err := compoundsFS.ReadFile("data/compounds.txt")
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
		"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/resource/ga/compounds.txt")
	want, err := os.ReadFile(official)
	if err != nil {
		dir := filepath.Dir(thisFile)
		var found []byte
		for i := 0; i < 12; i++ {
			candidate := filepath.Join(dir,
				"inspiration/languagetool/languagetool-language-modules/ga/src/main/resources/org/languagetool/resource/ga/compounds.txt")
			b, e := os.ReadFile(candidate)
			if e == nil {
				found = b
				break
			}
			dir = filepath.Dir(dir)
		}
		require.NotNil(t, found, "could not locate official ga/compounds.txt under inspiration/")
		want = found
		err = nil
	}
	require.NoError(t, err)
	require.True(t, bytes.Equal(want, embedded),
		"embedded ga/data/compounds.txt must be byte-identical to official Java resource")

	// Parse twin: both dash-only entries from official file
	d, err := rules.NewCompoundRuleData(bytes.NewReader(embedded), CompoundFilePath)
	require.NoError(t, err)
	require.True(t, d.ContainsIncorrect("mí úsáid"))
	require.True(t, d.ContainsDash("mí úsáid"))
	require.False(t, d.ContainsJoined("mí úsáid"))
	require.True(t, d.ContainsIncorrect("ró úsáid"))
	require.True(t, d.ContainsDash("ró úsáid"))
	require.False(t, d.ContainsJoined("ró úsáid"))
}
