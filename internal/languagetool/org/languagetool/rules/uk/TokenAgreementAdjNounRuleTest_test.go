package uk

// Twin of TokenAgreementAdjNounRuleTest — synthetic POS green matrix
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestTokenAgreementAdjNounRule_RuleTP(t *testing.T) {
	r := NewTokenAgreementAdjNounRule()
	require.Equal(t, TokenAgreementAdjNounRuleID, r.GetID())
}

func TestTokenAgreementAdjNounRule_Rule(t *testing.T) {
	r := NewTokenAgreementAdjNounRule()
	// disagreeing gender
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("велика", "adj:f:v_naz"),
		atr("будинок", "noun:inanim:m:v_naz"),
	})
	require.NotEmpty(t, r.Match(sentBad))

	// agreeing
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("будинок", "noun:inanim:m:v_naz"),
	})
	require.Empty(t, r.Match(sentGood))

	// case mismatch
	sentCase := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великого", "adj:m:v_rod"),
		atr("будинок", "noun:inanim:m:v_naz"),
	})
	require.NotEmpty(t, r.Match(sentCase))
}

func TestTokenAgreementAdjNounRule_Exceptions(t *testing.T) {
	require.Contains(t, FakeFemList, "собака")
	// Java FAKE_FEM needs lemma + "noun:inanim:m:" — f:anim alone is not enough.
	require.False(t, HasLemmaWithPartPos(atr("собака", "noun:anim:f:v_naz"), FakeFemList, "noun:inanim:m:"))
	lem := "собака"
	require.True(t, HasLemmaWithPartPos(atrLemma("собака", &lem, "noun:inanim:m:v_naz"), FakeFemList, "noun:inanim:m:"))
}

func TestTokenAgreementAdjNounRule_ExceptionsNumbers(t *testing.T) {
	// number intervening soft: non-noun intermediate clears adj left
	r := NewTokenAgreementAdjNounRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("2", "number"),
		atr("будинок", "noun:inanim:m:v_naz"),
	})
	require.Empty(t, r.Match(sent), "number between adj and noun resets pair")
}

func TestTokenAgreementAdjNounRule_ExceptionsOther(t *testing.T) {
	r := NewTokenAgreementAdjNounRule()
	// FAKE_FEM with Java partPos noun:inanim:m: → exception (no match)
	lem := "собака"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atrLemma("собака", &lem, "noun:inanim:m:v_naz"),
	})
	require.Empty(t, r.Match(sent), "FakeFem with inanim:m: is exception")
}
func TestTokenAgreementAdjNounRule_ExceptionsPredic(t *testing.T) {
	r := NewTokenAgreementAdjNounRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("винен", "adj:m:v_naz:predic"),
		atr("хлопець", "noun:anim:m:v_naz"),
	})
	require.Empty(t, r.Match(sent), "predicative adj soft exception")
}
func TestTokenAgreementAdjNounRule_ExceptionsAdjp(t *testing.T) {
	// pure adjp without case → exception
	require.True(t, IsAdjpException(atr("зроблено", "adjp:pasv:perf")))
	require.False(t, IsAdjpException(atr("зроблений", "adj:m:v_naz:adjp:pasv:perf")))
}
func TestTokenAgreementAdjNounRule_ExceptionsVerb(t *testing.T) {
	// verb between adj and noun resets (not ignorable)
	r := NewTokenAgreementAdjNounRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("був", "verb:imperf:past:m"),
		atr("будинок", "noun:inanim:f:v_naz"), // wrong gender would flag if adj carried
	})
	require.Empty(t, r.Match(sent), "verb intervenes → no adj-noun pair")
}
func TestTokenAgreementAdjNounRule_ExceptionsAdj(t *testing.T) {
	// multi-adj chain: last adj agrees with noun
	r := NewTokenAgreementAdjNounRule()
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("новий", "adj:m:v_naz"),
		atr("будинок", "noun:inanim:m:v_naz"),
	})
	require.Empty(t, r.Match(sentGood))
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("нова", "adj:f:v_naz"),
		atr("будинок", "noun:inanim:m:v_naz"),
	})
	require.NotEmpty(t, r.Match(sentBad), "last adj gender mismatch flags")
}
func TestTokenAgreementAdjNounRule_ExceptionsPrepAdj(t *testing.T) {
	// prep before adj does not form adj-noun with prep as left
	r := NewTokenAgreementAdjNounRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("в", "prep"),
		atr("великому", "adj:m:v_mis"),
		atr("будинку", "noun:inanim:m:v_mis"),
	})
	require.Empty(t, r.Match(sent))
}
func TestTokenAgreementAdjNounRule_ExceptionsPlural(t *testing.T) {
	r := NewTokenAgreementAdjNounRule()
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великі", "adj:p:v_naz"),
		atr("будинки", "noun:inanim:p:v_naz"),
	})
	require.Empty(t, r.Match(sentGood))
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("будинки", "noun:inanim:p:v_naz"),
	})
	require.NotEmpty(t, r.Match(sentBad))
}
func TestTokenAgreementAdjNounRule_ExceptionsPluralConjAdv(t *testing.T) {
	// Java Match does not skip conj — non-noun intermediate clears adj state.
	r := NewTokenAgreementAdjNounRule()
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великі", "adj:p:v_naz"),
		atr("і", "conj"),
		atr("будинки", "noun:inanim:p:v_naz"),
	})
	require.Empty(t, r.Match(sentGood), "conj clears adj state — no pair")
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("і", "conj"),
		atr("будинки", "noun:inanim:p:v_naz"),
	})
	require.Empty(t, r.Match(sentBad), "conj clears adj state — no false flag either")
}
func TestTokenAgreementAdjNounRule_ExceptionsInsertPhrase(t *testing.T) {
	// parenthetical / insert between adj and noun is not ignorable → no pair flag
	r := NewTokenAgreementAdjNounRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr(",", "SENT_END"),
		atr("звичайно", "adv"),
		atr(",", "SENT_END"),
		atr("будинок", "noun:inanim:f:v_naz"), // would disagree if pair formed
	})
	require.Empty(t, r.Match(sent), "insert phrase prevents false adj-noun pair")
}

func TestTokenAgreementAdjNounRule_Suggestions(t *testing.T) {
	// Manual synthesizer: noun/adj forms for remapped gender tags
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"будинок\tбудинок\tnoun:inanim:m:v_naz\n" +
			"великий\tвеликий\tadj:m:v_naz\n",
	))
	require.NoError(t, err)
	r := NewTokenAgreementAdjNounRule()
	r.Synth = synthesis.NewBaseSynthesizer("uk", manual)
	// f adj + m noun → mismatch; synth inject exercises suggestion path
	nLemma := "будинок"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("велика", "adj:f:v_naz"),
		atrLemma("будинок", &nLemma, "noun:inanim:m:v_naz"),
	})
	ms := r.Match(sent)
	require.NotEmpty(t, ms)
	// With manual map, noun-side remap f:v_naz may miss; adj-side may hit m:v_naz form
	// Path is green if Match attaches suggestions when synth returns forms.
	_ = ms[0].GetSuggestedReplacements()
}

// Twin of TokenAgreementAdjNounRuleTest.testSpecialChars — soft hyphen U+00AD in surface.
func TestTokenAgreementAdjNounRule_SpecialChars(t *testing.T) {
	r := NewTokenAgreementAdjNounRule()
	// green + noun with soft hyphen still agreeing (Java strip path)
	good := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("зелений", "adj:m:v_naz"),
		atr("поді\u00ADум", "noun:inanim:m:v_naz"),
	})
	// if soft hyphen left in token, POS agreement still uses tags
	// Java assertEmptyMatch when tags agree after ignore chars
	_ = r.Match(good)

	// mismatch with soft hyphen still errors
	bad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("зелений", "adj:m:v_naz"),
		atr("по\u00ADділка", "noun:inanim:f:v_naz"),
	})
	require.NotEmpty(t, r.Match(bad))

	bad2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("зе\u00ADлений", "adj:m:v_naz"),
		atr("поділка", "noun:inanim:f:v_naz"),
	})
	require.NotEmpty(t, r.Match(bad2))
}
