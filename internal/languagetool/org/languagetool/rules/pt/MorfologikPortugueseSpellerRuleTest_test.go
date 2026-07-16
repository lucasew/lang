package pt

// Twin of MorfologikPortugueseSpellerRuleTest — map speller surface.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func withPTSpeller(words ...string) *MorfologikPortugueseSpellerRule {
	sp := morfologik.NewMorfologikSpeller(PortuguesePTDict, 1)
	for _, w := range words {
		sp.AddWord(w)
	}
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	return r
}

func TestMorfologikPortugueseSpeller_PortugueseSpellerSanity(t *testing.T) {
	r := withPTSpeller("casa", "teste")
	require.Equal(t, MorfologikPortuguesePTSpellerRuleID, r.GetID())
	require.False(t, r.Speller.IsMisspelled("casa"))
	require.True(t, r.Speller.IsMisspelled("caza"))
}

func TestMorfologikPortugueseSpeller_PortugueseSpellerSpecificIds(t *testing.T) {
	require.Equal(t, "MORFOLOGIK_RULE_PT_PT", NewMorfologikPortugalPortugueseSpellerRule().GetID())
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR", NewMorfologikBrazilianPortugueseSpellerRule().GetID())
}

func TestMorfologikPortugueseSpeller_EuropeanPortugueseSpelling(t *testing.T) {
	r := withPTSpeller("facto")
	sent := languagetool.AnalyzePlain("fcto")
	matches, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, matches, 1)
}

func TestMorfologikPortugueseSpeller_AfricanPortugueseSpelling(t *testing.T) {
	r := NewMorfologikPortugueseSpellerRule("pt-AO", "/pt/hunspell/pt_AO.dict", "MORFOLOGIK_RULE_PT_AO")
	require.Equal(t, "pt-AO", r.VariantCode)
}

func TestMorfologikPortugueseSpeller_BrazilianPortugueseSpelling(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	sp.AddWord("fato")
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.False(t, r.Speller.IsMisspelled("fato"))
}

func TestMorfologikPortugueseSpeller_EuropeanPortugueseHyphenatedClitics(t *testing.T) {
	r := withPTSpeller("dá-se")
	require.False(t, r.Speller.IsMisspelled("dá-se"))
}

func TestMorfologikPortugueseSpeller_BrazilianPortugueseHyphenatedClitics(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	sp.AddWord("dá-se")
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	require.False(t, r.Speller.IsMisspelled("dá-se"))
}

func TestMorfologikPortugueseSpeller_PortugueseSpellerDoesNotAcceptVerbFormsWithElidedConsonants(t *testing.T) {
	r := withPTSpeller("estar")
	require.True(t, r.Speller.IsMisspelled("tar")) // not in dict
}

func TestMorfologikPortugueseSpeller_PortugueseSpellerAcceptsVerbsWithProductivePrefixes(t *testing.T) {
	r := withPTSpeller("recomeçar")
	require.False(t, r.Speller.IsMisspelled("recomeçar"))
}

func TestMorfologikPortugueseSpeller_PortugueseHyphenationRules(t *testing.T) {
	r := withPTSpeller("guarda-chuva")
	require.False(t, r.Speller.IsMisspelled("guarda-chuva"))
}

func TestMorfologikPortugueseSpeller_PortugueseSymmetricalDialectDifferences(t *testing.T) {
	// PT accepts facto; BR accepts fato — different variants.
	pt := withPTSpeller("facto")
	br := NewMorfologikBrazilianPortugueseSpellerRule()
	brSp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	brSp.AddWord("fato")
	br.Speller = brSp
	br.IsMisspelled = brSp.IsMisspelled
	require.False(t, pt.Speller.IsMisspelled("facto"))
	require.True(t, pt.Speller.IsMisspelled("fato"))
	require.False(t, br.Speller.IsMisspelled("fato"))
}

func TestMorfologikPortugueseSpeller_PortugueseAsymmetricalDialectDifferences(t *testing.T) {
	pt := withPTSpeller("óleo")
	require.False(t, pt.Speller.IsMisspelled("óleo"))
}
