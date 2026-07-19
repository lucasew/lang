package uk

// Twin of TokenAgreementAdjNounRuleTest — synthetic POS green matrix
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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
	// conj "і" is ignorable intervening; last adj… actually adj then і then noun
	r := NewTokenAgreementAdjNounRule()
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великі", "adj:p:v_naz"),
		atr("і", "conj"),
		atr("будинки", "noun:inanim:p:v_naz"),
	})
	require.Empty(t, r.Match(sentGood), "conj between agreeing adj-noun passes")
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("великий", "adj:m:v_naz"),
		atr("і", "conj"),
		atr("будинки", "noun:inanim:p:v_naz"),
	})
	require.NotEmpty(t, r.Match(sentBad), "conj does not hide number mismatch")
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
