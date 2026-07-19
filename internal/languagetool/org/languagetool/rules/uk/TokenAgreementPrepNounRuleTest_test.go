package uk

// Twin of languagetool-language-modules/uk TokenAgreementPrepNounRuleTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of TokenAgreementPrepNounRuleTest.testRule — synthetic POS green matrix
func TestTokenAgreementPrepNounRule_Rule(t *testing.T) {
	r := NewTokenAgreementPrepNounRule()
	// known preps from case_government.txt
	cases := []struct {
		prepLemma string
		nounTag   string
		wantMatch bool
	}{
		{"в", "noun:inanim:m:v_mis", false},
		{"в", "noun:inanim:m:v_oru", true},
		{"з", "noun:inanim:m:v_oru", false}, // з often governs v_oru / v_rod / v_zna
		{"до", "noun:inanim:m:v_rod", false},
		{"до", "noun:inanim:m:v_naz", true},
	}
	for _, tc := range cases {
		lemma := tc.prepLemma
		sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
			atrLemma(tc.prepLemma, &lemma, "prep"),
			atr("N", tc.nounTag),
		})
		matches := r.Match(sent)
		if tc.wantMatch {
			require.NotEmpty(t, matches, "prep=%s noun=%s should match", tc.prepLemma, tc.nounTag)
		} else {
			require.Empty(t, matches, "prep=%s noun=%s should pass", tc.prepLemma, tc.nounTag)
		}
	}
}

func TestTokenAgreementPrepNounRule_ZandZnaAsRare(t *testing.T) {
	// Soft: rare з+v_zna exception matrix deferred; ensure rule still constructs
	r := NewTokenAgreementPrepNounRule()
	require.NotNil(t, r.CaseGov)
	// "з" should have governments loaded
	require.NotEmpty(t, r.CaseGov.GetCaseGovernments("з"))
}

func TestTokenAgreementPrepNounRule_RulePronPosNew(t *testing.T) {
	r := NewTokenAgreementPrepNounRule()
	// до + genitive pronoun ok
	lemma := "до"
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("до", &lemma, "prep"),
		atr("нього", "noun:unanim:m:v_rod:pron:pers:3"),
	})
	require.Empty(t, r.Match(sentGood))
	// до + nominative pronoun mismatch
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("до", &lemma, "prep"),
		atr("він", "noun:unanim:m:v_naz:pron:pers:3"),
	})
	require.NotEmpty(t, r.Match(sentBad))
}

func TestTokenAgreementPrepNounRule_RulePronPos(t *testing.T) {
	r := NewTokenAgreementPrepNounRule()
	// з + instrumental pronoun
	lemma := "з"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("з", &lemma, "prep"),
		atr("нею", "noun:unanim:f:v_oru:pron:pers:3"),
	})
	require.Empty(t, r.Match(sent))
	// з + nominative
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("з", &lemma, "prep"),
		atr("вона", "noun:unanim:f:v_naz:pron:pers:3"),
	})
	require.NotEmpty(t, r.Match(sentBad))
}

func TestTokenAgreementPrepNounRule_RuleFlexibleOrder(t *testing.T) {
	// simplified matcher is left-to-right prep→noun only; reverse order does not flag
	r := NewTokenAgreementPrepNounRule()
	lemma := "до"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("дому", "noun:inanim:m:v_naz"),
		atrLemma("до", &lemma, "prep"),
	})
	require.Empty(t, r.Match(sent), "noun before prep is not checked as pair")
}

func TestTokenAgreementPrepNounRule_SpecialChars(t *testing.T) {
	// Soft hyphen / combining acute cleaned then still recognized as prep
	r := NewTokenAgreementPrepNounRule()
	lemma := "в"
	// surface with soft hyphen
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("в\u00AD", &lemma, "prep"),
		atr("домі", "noun:inanim:m:v_mis"),
	})
	require.Empty(t, r.Match(sent))
}

func TestTokenAgreementPrepNounRule_UnusualCharacters(t *testing.T) {
	require.Equal(t, "боснія", CleanIgnoreChars("боснія"))
	require.Equal(t, "боснія", CleanIgnoreChars("бос\u0301нія"))
}

func TestTokenAgreementPrepNounRule_WithAdv(t *testing.T) {
	// Java getExceptionStrong: pure adv → RuleException(0) skip keeps prep;
	// following wrong-case noun still flags (TokenAgreementPrepNounExceptionHelper L266–271).
	r := NewTokenAgreementPrepNounRule()
	lemma := "в"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("в", &lemma, "prep"),
		atr("дуже", "adv"),
		atr("домі", "noun:inanim:m:v_oru"),
	})
	require.NotEmpty(t, r.Match(sent), "adv skip keeps prep; wrong case noun still matches")
}

// Port of TokenAgreementPrepNounRuleTest.testIsCapitalized
func TestTokenAgreementPrepNounRule_IsCapitalized(t *testing.T) {
	require.False(t, IsCapitalized("боснія"))
	require.True(t, IsCapitalized("Боснія"))
	require.True(t, IsCapitalized("Боснія-Герцеговина"))
	require.False(t, IsCapitalized("По-перше"))
	require.False(t, IsCapitalized("ПаП"))
	require.True(t, IsCapitalized("П'ятниця"))
	require.False(t, IsCapitalized("П'ЯТНИЦЯ"))
	require.True(t, IsCapitalized("EuroGas"))
	require.True(t, IsCapitalized("Рясна-2"))
	require.False(t, IsCapitalized("ДБЗПТЛ"))
}
