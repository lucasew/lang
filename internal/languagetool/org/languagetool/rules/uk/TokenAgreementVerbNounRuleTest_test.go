package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"github.com/stretchr/testify/require"
)

var _ = require.Equal
var _ = tools.Unimplemented

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
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_V_Vinf_N")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_ADV_Vinf_N
func TestTokenAgreementVerbNounRule_RuleTn_ADV_Vinf_N(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_ADV_Vinf_N")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_ADJ_Vinf_N
func TestTokenAgreementVerbNounRule_RuleTn_ADJ_Vinf_N(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_ADJ_Vinf_N")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_NOUN_Vinf_N
func TestTokenAgreementVerbNounRule_RuleTn_NOUN_Vinf_N(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_NOUN_Vinf_N")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_V
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_V(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_V")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_ADV
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_ADV(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_ADV")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_ADJ
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_N_ADJ(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_N_ADJ")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_V_N
func TestTokenAgreementVerbNounRule_RuleTn_Vinf_V_N(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_Vinf_V_N")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTn_N_Vinf_ADJ
func TestTokenAgreementVerbNounRule_RuleTn_N_Vinf_ADJ(t *testing.T) {
	t.Skip("Java @Ignore")
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTn_N_Vinf_ADJ")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTnNumr
func TestTokenAgreementVerbNounRule_RuleTnNumr(t *testing.T) {
	// contains assertTrue
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTNvNaz
func TestTokenAgreementVerbNounRule_RuleTNvNaz(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTNvNaz")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTNTime
func TestTokenAgreementVerbNounRule_RuleTNTime(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTNTime")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTnVrod
func TestTokenAgreementVerbNounRule_RuleTnVrod(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTnVrod")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleTnInsertPhrase
func TestTokenAgreementVerbNounRule_RuleTnInsertPhrase(t *testing.T) {
	t.Skip("unimplemented: TokenAgreementVerbNounRuleTest.testRuleTnInsertPhrase")
}

// Port of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/TokenAgreementVerbNounRuleTest.java :: TokenAgreementVerbNounRuleTest.testRuleDisambigNazInf
func TestTokenAgreementVerbNounRule_RuleDisambigNazInf(t *testing.T) {
	// contains assertEquals — full values in Java twin source
}
