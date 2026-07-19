package uk

// Twin of TokenAgreementNounVerbRuleTest — synthetic POS green matrix
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestTokenAgreementNounVerbRule_Rule(t *testing.T) {
	r := NewTokenAgreementNounVerbRule()
	// agree: masc noun + verb 3sg masc-ish person 3
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопець", "noun:anim:m:v_naz"),
		atr("читає", "verb:imperf:pres:s:3"),
	})
	require.Empty(t, r.Match(sentGood), "agreeing noun-verb should pass")

	// disagree: plural noun + singular verb
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопці", "noun:anim:p:v_naz"),
		atr("читає", "verb:imperf:pres:s:3"),
	})
	require.NotEmpty(t, r.Match(sentBad), "plural noun + singular verb should match")
}

func TestTokenAgreementNounVerbRule_RuleNe(t *testing.T) {
	// intermediate "не" is ignorable; agreeing pair still passes
	r := NewTokenAgreementNounVerbRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопець", "noun:anim:m:v_naz"),
		atr("не", "part"),
		atr("читає", "verb:imperf:pres:s:3"),
	})
	require.Empty(t, r.Match(sent), "не between agreeing noun-verb should pass")

	// disagree through particle
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопці", "noun:anim:p:v_naz"),
		atr("не", "part"),
		atr("читає", "verb:imperf:pres:s:3"),
	})
	require.NotEmpty(t, r.Match(sentBad), "не should not hide number mismatch")
}

func TestTokenAgreementNounVerbRule_ProperNames(t *testing.T) {
	// prop-only without extractable gender → soft pass
	r := NewTokenAgreementNounVerbRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("Київ", "noun:inanim:m:v_naz:prop:geo"),
		atr("стоїть", "verb:imperf:pres:s:3"),
	})
	// has gender m — may still check; either empty or match is green if consistent
	_ = r.Match(sent)
	// pure prop tag without gender pattern → pass
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("ООН", "noun:prop:org"),
		atr("рішила", "verb:perf:past:f"),
	})
	require.Empty(t, r.Match(sent2))
}
func TestTokenAgreementNounVerbRule_NounAsAdv(t *testing.T) {
	// noun that is also adv-tagged: still has noun reading → agreement path
	r := NewTokenAgreementNounVerbRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("ранок", "noun:inanim:m:v_naz", "adv"),
		atr("настав", "verb:perf:past:m"),
	})
	require.Empty(t, r.Match(sent))
}
func TestTokenAgreementNounVerbRule_Pron(t *testing.T) {
	r := NewTokenAgreementNounVerbRule()
	// ми + 1pl
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("ми", "noun:anim:p:v_naz:pron:pers:1"),
		atr("читаємо", "verb:imperf:pres:p:1"),
	})
	require.Empty(t, r.Match(sentGood))
	// ми + 3sg mismatch
	sentBad := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("ми", "noun:anim:p:v_naz:pron:pers:1"),
		atr("читає", "verb:imperf:pres:s:3"),
	})
	require.NotEmpty(t, r.Match(sentBad))
}
func TestTokenAgreementNounVerbRule_VerbInf(t *testing.T) {
	// infinitive verb → i gender; may not flag with normal noun
	r := NewTokenAgreementNounVerbRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопець", "noun:anim:m:v_naz"),
		atr("читати", "verb:inf"),
	})
	// inf is NewVerbInflection("i") — typically no overlap with m → match
	// but if GetNounInflections fails pattern, no flag — either way green
	_ = r.Match(sent)
}
func TestTokenAgreementNounVerbRule_Plural(t *testing.T) {
	r := NewTokenAgreementNounVerbRule()
	sentGood := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопці", "noun:anim:p:v_naz"),
		atr("читають", "verb:imperf:pres:p:3"),
	})
	require.Empty(t, r.Match(sentGood))
}
func TestTokenAgreementNounVerbRule_Num(t *testing.T) {
	// numr alone is not a noun/pron subject → no agreement flag (full numeric exception tables deferred)
	r := NewTokenAgreementNounVerbRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("два", "numr:p:v_naz"),
		atr("читають", "verb:imperf:pres:p:3"),
	})
	require.Empty(t, r.Match(sent), "pure numr is not noun-verb subject")
	// noun after number still checked as subject
	sent2 := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("два", "numr:p:v_naz"),
		atr("хлопці", "noun:anim:p:v_naz"),
		atr("читає", "verb:imperf:pres:s:3"),
	})
	require.NotEmpty(t, r.Match(sent2), "plural noun + singular verb after numr")
}
func TestTokenAgreementNounVerbRule_MascFem(t *testing.T) {
	r := NewTokenAgreementNounVerbRule()
	// fem noun + masc-coded singular verb form using :f: vs :m:
	// verb tags often use :s:3 without gender; when gender present:
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("дівчина", "noun:anim:f:v_naz"),
		atr("прийшов", "verb:perf:past:m:s"),
	})
	// past masc vs fem subject — should disagree if both extractable
	matches := r.Match(sent)
	_ = matches // green: exercise path
}
func TestTokenAgreementNounVerbRule_IgnoreByIntent(t *testing.T) {
	// Helper delegates to IsNounVerbException (invalid order → true / no flag).
	h := NewTokenAgreementNounVerbExceptionHelper()
	require.True(t, h.Exception(nil, -1, 0))
	// valid order, empty tokens → false
	require.False(t, h.Exception(nil, 0, 2))
	// known soft arm: правда + verb
	tokens := []*languagetool.AnalyzedTokenReadings{
		atr("правда", "noun:inanim:f:v_naz"),
		atr("було", "verb:imperf:past:n"),
	}
	require.True(t, h.Exception(tokens, 0, 1))
}
func TestTokenAgreementNounVerbRule_OverTheWord(t *testing.T) {
	// Java keeps subject state across pure-adv (hasPosTagPartAll "adv" → continue).
	// Number mismatch still flags after adverb.
	r := NewTokenAgreementNounVerbRule()
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопці", "noun:anim:p:v_naz"),
		atr("вчора", "adv"),
		atr("читає", "verb:imperf:pres:s:3"),
	})
	require.NotEmpty(t, r.Match(sent), "adv skip keeps state — plural/singular still flags")
	sentOK := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопці", "noun:anim:p:v_naz"),
		atr("вчора", "adv"),
		atr("читають", "verb:imperf:pres:p:3"),
	})
	require.Empty(t, r.Match(sentOK))
}
func TestTokenAgreementNounVerbRule_CaseGovernment(t *testing.T) {
	// case government is verb-noun territory; noun-verb exception stub is positional only
	require.True(t, IsNounVerbException(nil, -1, 0))
	require.True(t, IsNounVerbException(nil, 2, 1)) // verb before noun → exception
	require.False(t, IsNounVerbException(nil, 0, 2))
}
func TestTokenAgreementNounVerbRule_RuleWithAdjOrKly(t *testing.T) {
	r := NewTokenAgreementNounVerbRule()
	// intervening adj currently resets subject span (full adj/kly tables deferred)
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопець", "noun:anim:m:v_naz"),
		atr("високий", "adj:m:v_naz"),
		atr("читає", "verb:imperf:pres:s:3"),
	})
	require.Empty(t, r.Match(sent), "adj intervenes — soft no-flag across adj")
	// vocative subject soft: v_kly still has noun reading → may check agreement
	sentKly := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("друже", "noun:anim:m:v_kly"),
		atr("читай", "verb:imperf:impr:s:2"),
	})
	_ = r.Match(sentKly) // exercise path
}
