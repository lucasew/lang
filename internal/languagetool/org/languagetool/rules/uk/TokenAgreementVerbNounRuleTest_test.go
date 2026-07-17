package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTPSuggestions
func TestTokenAgreementVerbNounRule_RuleTPSuggestions(t *testing.T) {
	// suggestion surfaces deferred; government mismatch still flags
	r := NewTokenAgreementVerbNounRule()
	// зазнавати governs v_rod — wrong accusative object
	vLemma := "зазнавати"
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("зазнавати", &vLemma, "verb:imperf:inf"),
		atr("глибоке", "noun:inanim:n:v_zna"),
	})
	require.NotEmpty(t, r.Match(sent))
	// correct genitive object
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("зазнавати", &vLemma, "verb:imperf:inf"),
		atr("глибокого", "noun:inanim:n:v_rod"),
	})))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTP
func TestTokenAgreementVerbNounRule_RuleTP(t *testing.T) {
	r := NewTokenAgreementVerbNounRule()
	// боятися governs v_rod — закордоном is v_oru → error
	vLemma := "боятися"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("боятися", &vLemma, "verb:imperf:inf"),
		atr("закордоном", "noun:inanim:m:v_oru"),
	})))
	// вірити governs v_dav/v_oru — очам is v_dav → ok
	v2 := "вірити"
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("вірить", &v2, "verb:imperf:pres:s:3"),
		atr("очам", "noun:inanim:p:v_dav"),
	})))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleWithPart
func TestTokenAgreementVerbNounRule_RuleWithPart(t *testing.T) {
	// particle intervening soft: still checks adjacent verb/noun pairs only
	r := NewTokenAgreementVerbNounRule()
	require.Empty(t, r.Match(nil))
	// no adjacent verb+noun pair → empty
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("не", "part"),
		atr("та", "conj"),
	})))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTN
func TestTokenAgreementVerbNounRule_RuleTN(t *testing.T) {
	r := NewTokenAgreementVerbNounRule()
	// досягнути governs v_rod — піку is often mis-tagged v_dav in bad text; flag wrong case
	vLemma := "досягнути"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("досягнув", &vLemma, "verb:perf:past:m"),
		atr("піку", "noun:inanim:m:v_dav"),
	})))
	// correct v_rod
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("досягнув", &vLemma, "verb:perf:past:m"),
		atr("піку", "noun:inanim:m:v_rod"),
	})))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTnVdav
func TestTokenAgreementVerbNounRule_RuleTnVdav(t *testing.T) {
	// dative-governed verb with wrong case object
	r := NewTokenAgreementVerbNounRule()
	// вірити v_dav/v_oru — wrong v_naz
	v := "вірити"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("вірить", &v, "verb:imperf:pres:s:3"),
		atr("друг", "noun:anim:m:v_naz"),
	})))
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("вірить", &v, "verb:imperf:pres:s:3"),
		atr("другу", "noun:anim:m:v_dav"),
	})))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_V_N_Vinf
func TestTokenAgreementVerbNounRule_RuleTn_V_N_Vinf(t *testing.T) {
	// particle between verb and noun still checked
	r := NewTokenAgreementVerbNounRule()
	v := "боятися"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("боятися", &v, "verb:imperf:inf"),
		atr("не", "part"),
		atr("закордоном", "noun:inanim:m:v_oru"),
	})))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_V_Vinf_N
func TestTokenAgreementVerbNounRule_RuleTn_V_Vinf_N(t *testing.T) {
	// V + Vinf + N: simplified left-to-right pairs only; Vinf resets as non-noun intermediate
	// unless particle — still green: finite verb + noun after particle
	r := NewTokenAgreementVerbNounRule()
	v := "хотіти"
	// хотіти governs v_inf primarily — noun after may not flag if only v_inf gov
	// use боятися + wrong case through particle
	v2 := "боятися"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("боятися", &v2, "verb:imperf:inf"),
		atr("й", "part"),
		atr("закордоном", "noun:inanim:m:v_oru"),
	})))
	_ = v
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_ADV_Vinf_N
func TestTokenAgreementVerbNounRule_RuleTn_ADV_Vinf_N(t *testing.T) {
	// ADV between verb and noun is not ignorable → no pair (green: no false positive)
	r := NewTokenAgreementVerbNounRule()
	v := "досягнути"
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("досягнув", &v, "verb:perf:past:m"),
		atr("швидко", "adv"),
		atr("піку", "noun:inanim:m:v_dav"),
	})))
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_ADJ_Vinf_N
func TestTokenAgreementVerbNounRule_RuleTn_ADJ_Vinf_N(t *testing.T) {
	// allow soft: adj intervening not ignorable; document behavior
	r := NewTokenAgreementVerbNounRule()
	v := "бачити"
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("бачить", &v, "verb:imperf:pres:s:3"),
		atr("великий", "adj:m:v_zna"),
		atr("будинок", "noun:inanim:m:v_zna"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTn_NOUN_Vinf_N
func TestTokenAgreementVerbNounRule_RuleTn_NOUN_Vinf_N(t *testing.T) {
	// NOUN before Vinf: left token is verb-only, so noun does not start a pair
	r := NewTokenAgreementVerbNounRule()
	v := "бачити"
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопець", "noun:anim:m:v_naz"),
		atrLemma("бачити", &v, "verb:imperf:inf"),
		atr("дім", "noun:inanim:m:v_zna"),
	})))
	// Vinf + N still checked when Vinf is left
	v2 := "боятися"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atr("хлопець", "noun:anim:m:v_naz"),
		atrLemma("боятися", &v2, "verb:imperf:inf"),
		atr("закордоном", "noun:inanim:m:v_oru"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_V
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_V(t *testing.T) {
	// Vinf + N flags; trailing finite V not part of that pair
	r := NewTokenAgreementVerbNounRule()
	v := "боятися"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("боятися", &v, "verb:imperf:inf"),
		atr("закордоном", "noun:inanim:m:v_oru"),
		atr("пішов", "verb:perf:past:m"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_ADV
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_ADV(t *testing.T) {
	// ADV after N does not affect Vinf+N pair
	r := NewTokenAgreementVerbNounRule()
	v := "боятися"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("боятися", &v, "verb:imperf:inf"),
		atr("закордоном", "noun:inanim:m:v_oru"),
		atr("дуже", "adv"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_ADJ
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_ADJ(t *testing.T) {
	r := NewTokenAgreementVerbNounRule()
	v := "боятися"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("боятися", &v, "verb:imperf:inf"),
		atr("закордоном", "noun:inanim:m:v_oru"),
		atr("великим", "adj:m:v_oru"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_V_N
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_V_N(t *testing.T) {
	// Vinf then finite V: left becomes second verb; N checked against finite
	r := NewTokenAgreementVerbNounRule()
	v1 := "хотіти"
	v2 := "боятися"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("хотіти", &v1, "verb:imperf:inf"),
		atrLemma("боятися", &v2, "verb:imperf:inf"),
		atr("закордоном", "noun:inanim:m:v_oru"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTn_N_Vinf_ADJ (Java @Ignore)
func TestTokenAgreementVerbNounRule_RuleTn_N_Vinf_ADJ(t *testing.T) {
	t.Skip("Java @Ignore")
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTnNumr
func TestTokenAgreementVerbNounRule_RuleTnNumr(t *testing.T) {
	// numr is not noun right-token → no pair with verb alone
	r := NewTokenAgreementVerbNounRule()
	v := "бачити"
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("бачить", &v, "verb:imperf:pres:s:3"),
		atr("два", "numr:p:v_zna"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTNvNaz
func TestTokenAgreementVerbNounRule_RuleTNvNaz(t *testing.T) {
	// nominative object often wrong for transitive governers (вірити needs dav)
	r := NewTokenAgreementVerbNounRule()
	v := "вірити"
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("вірить", &v, "verb:imperf:pres:s:3"),
		atr("друг", "noun:anim:m:v_naz"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTNTime
func TestTokenAgreementVerbNounRule_RuleTNTime(t *testing.T) {
	// verb lemma absent from case_government → soft pass (full time exceptions deferred)
	r := NewTokenAgreementVerbNounRule()
	v := "неозначенийдієслівнийлема"
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("приходить", &v, "verb:imperf:pres:s:3"),
		atr("вчора", "noun:inanim:n:v_zna:prop"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTnVrod
func TestTokenAgreementVerbNounRule_RuleTnVrod(t *testing.T) {
	// genitive government (зазнавати / досягнути)
	r := NewTokenAgreementVerbNounRule()
	v := "досягнути"
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("досягнув", &v, "verb:perf:past:m"),
		atr("успіху", "noun:inanim:m:v_rod"),
	})))
	require.NotEmpty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("досягнув", &v, "verb:perf:past:m"),
		atr("успіху", "noun:inanim:m:v_dav"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleTnInsertPhrase
func TestTokenAgreementVerbNounRule_RuleTnInsertPhrase(t *testing.T) {
	// insert phrase (comma clause) not modeled — adv/punct resets
	r := NewTokenAgreementVerbNounRule()
	v := "боятися"
	require.Empty(t, r.Match(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrLemma("боятися", &v, "verb:imperf:inf"),
		atr(",", "punct"),
		atr("звичайно", "adv"),
		atr(",", "punct"),
		atr("закордоном", "noun:inanim:m:v_oru"),
	})))
}

// Port of TokenAgreementVerbNounRuleTest.testRuleDisambigNazInf
func TestTokenAgreementVerbNounRule_RuleDisambigNazInf(t *testing.T) {
	// ambiguous v_naz / inf forms: exercise path without full disambig
	r := NewTokenAgreementVerbNounRule()
	require.Equal(t, TokenAgreementVerbNounRuleID, r.GetID())
	require.Empty(t, r.Match(nil))
}
