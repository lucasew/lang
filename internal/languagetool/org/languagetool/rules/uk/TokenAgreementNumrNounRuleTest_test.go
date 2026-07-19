package uk

// Twin of TokenAgreementNumrNounRuleTest — synthetic POS green matrix
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestTokenAgreementNumrNounRule_Rule(t *testing.T) {
	r := NewTokenAgreementNumrNounRule()
	// agreeing numr+noun same case/gender
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("два", "numr:m:v_naz"),
		atr("дні", "noun:inanim:m:v_naz"),
	})
	require.Empty(t, r.Match(sentGood))

	// disagree: numr feminine vs noun masculine
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("дві", "numr:f:v_naz"),
		atr("дні", "noun:inanim:m:v_naz"),
	})
	require.NotEmpty(t, r.Match(sentBad))
}

func TestTokenAgreementNumrNounRule_RuleTN(t *testing.T) {
	// force construct + case government path for "тон" style numbers soft
	r := NewTokenAgreementNumrNounRule()
	require.Equal(t, TokenAgreementNumrNounRuleID, r.GetID())
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("три", "numr:m:v_naz"),
		atr("роки", "noun:inanim:m:v_naz"),
	})
	_ = r.Match(sent)
}

func TestTokenAgreementNumrNounRule_RuleForceNoun(t *testing.T) {
	r := NewTokenAgreementNumrNounRule()
	// Java always flags surface «тон» → suggest «тонн» (before force-pattern agreement path)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("дві", "numr:f:v_naz"),
		atr("тон", "noun:inanim:m:v_naz"),
	})
	require.NotEmpty(t, r.Match(sent), "тон surface suggests тонн")
	// force-noun unit (кілобайт) still available via IsForceNounException helper
	require.True(t, IsForceNounException(nil, atr("кілобайт", "noun:inanim:m:v_naz")))
}
func TestTokenAgreementNumrNounRule_RuleTon(t *testing.T) {
	// Java NOUN_FORCE_PATTERN is Matcher.matches() full-string — "тон" yes, "тони" no invent.
	require.True(t, IsForceNounException(nil, atr("тон", "noun:inanim:m:v_naz")))
	require.False(t, IsForceNounException(nil, atr("тони", "noun:inanim:p:v_naz")))
	require.False(t, IsForceNounException(nil, atr("дні", "noun:inanim:m:v_naz")))
	// SI/measure units from Java pattern
	require.True(t, IsForceNounException(nil, atr("кілобайт", "noun:inanim:m:v_naz")))
	require.True(t, IsForceNounException(nil, atr("чоловік", "noun:anim:m:v_naz")))
}
func TestTokenAgreementNumrNounRule_RuleFract(t *testing.T) {
	r := NewTokenAgreementNumrNounRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("півтора", "numr:n:v_naz"),
		atr("року", "noun:inanim:m:v_rod"),
	})
	require.Empty(t, r.Match(sent), "fractional numr exception")
}
func TestTokenAgreementNumrNounRule_RuleFractionals(t *testing.T) {
	require.True(t, IsFractionalNumrException(atr("півтори", "numr"), atr("години", "noun:f:v_naz")))
	require.False(t, IsFractionalNumrException(atr("три", "numr:m:v_naz"), atr("дні", "noun:m:v_naz")))
}
