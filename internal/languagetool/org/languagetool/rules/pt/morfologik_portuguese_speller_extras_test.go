package pt

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	taggingpt "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/pt"
	"github.com/stretchr/testify/require"
)

func TestCheckDiaeresis(t *testing.T) {
	// only ü path (Java replace ü → u)
	require.Equal(t, "linguistica", checkDiaeresis("lingüistica"))
	require.Empty(t, checkDiaeresis("casa"))
	require.Empty(t, checkDiaeresis("teste"))
}

func TestCheckEuropeanStyle1PLPastTense(t *testing.T) {
	require.Equal(t, "falamos", checkEuropeanStyle1PLPastTense("pt-BR", "falámos"))
	require.Empty(t, checkEuropeanStyle1PLPastTense("pt-PT", "falámos"))
	require.Empty(t, checkEuropeanStyle1PLPastTense("pt-BR", "falamos"))
}

func TestIsTitlecasedHyphenatedWord(t *testing.T) {
	require.True(t, isTitlecasedHyphenatedWord([]string{"Guarda", "Chuva"}))
	require.True(t, isTitlecasedHyphenatedWord([]string{"GUARDA", "CHUVA"}))
	require.False(t, isTitlecasedHyphenatedWord([]string{"Guarda", "cHuva"})) // mixed
}

func TestDialectAlternationMapping_BR(t *testing.T) {
	m := loadDialectAlternationMapping("pt-BR")
	if len(m) == 0 {
		t.Skip("dialect_alternations.txt not discoverable")
	}
	// European form keys → Brazilian suggestion (file Abdônio=Abdónio)
	require.Equal(t, "Abdônio", m["abdónio"])
}

func TestDialectAlternationMapping_PT(t *testing.T) {
	m := loadDialectAlternationMapping("pt-PT")
	if len(m) == 0 {
		t.Skip("dialect_alternations.txt not discoverable")
	}
	require.Equal(t, "Abdónio", m["abdônio"])
}

func TestFilterDoNotSuggest(t *testing.T) {
	// load real list when present
	_ = getDoNotSuggestWords()
	sugs := filterDoNotSuggest([]string{"casa", "aidético", "teste"})
	// if list loaded, aidético dropped
	if len(getDoNotSuggestWords()) > 0 {
		require.NotContains(t, sugs, "aidético")
		require.Contains(t, sugs, "casa")
	}
}

func TestMatch_TitlecaseHyphenAcceptedWhenLowerOK(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PortuguesePTDict, 1)
	sp.AddWord("guarda-chuva")
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// Titlecase each part, lower form in dict → no match
	sent := languagetool.AnalyzePlain("Guarda-Chuva")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestMatch_DiaeresisSuggestion(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PortuguesePTDict, 1)
	// no words → empty dict fail-closed won't flag; inject misspell path
	sp.AddWord("casa")
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// word with ü not in dict → match with ü→u suggestion
	sent := languagetool.AnalyzePlain("lingüistica")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Equal(t, []string{"linguistica"}, m[0].GetSuggestedReplacements())
	require.Contains(t, m[0].GetMessage(), "trema")
}

func TestMatch_BRPastTense(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	sp.AddWord("casa")
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("falámos")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Equal(t, []string{"falamos"}, m[0].GetSuggestedReplacements())
}

func TestMatch_DialectSurface(t *testing.T) {
	if len(loadDialectAlternationMapping("pt-BR")) == 0 {
		t.Skip("no dialect map")
	}
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	sp.AddWord("casa")
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// European form flagged on BR with dialect suggestion
	sent := languagetool.AnalyzePlain("Abdónio")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Equal(t, []string{"Abdônio"}, m[0].GetSuggestedReplacements())
	require.Contains(t, m[0].GetMessage(), "europeu")
}

func TestIsValidCliticVerb_FailClosedWithoutTagger(t *testing.T) {
	r := NewMorfologikPortugalPortugueseSpellerRule()
	require.False(t, r.isValidCliticVerb("diz-se"))
	r.TagPOS = func(word string) []string {
		if word == "diz-se" {
			return []string{"VMIP3S0:P0"}
		}
		return nil
	}
	require.True(t, r.isValidCliticVerb("diz-se"))
}

func TestGetIdForDialectIssue(t *testing.T) {
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR_DIALECT", NewMorfologikBrazilianPortugueseSpellerRule().getIdForDialectIssue())
	require.Equal(t, "MORFOLOGIK_RULE_PT_PT_DIALECT", NewMorfologikPortugalPortugueseSpellerRule().getIdForDialectIssue())
}

func TestMatch_DialectSetsSpecificRuleId(t *testing.T) {
	if len(loadDialectAlternationMapping("pt-BR")) == 0 {
		t.Skip("no dialect map")
	}
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	sp.AddWord("casa")
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("Abdónio")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR_DIALECT", m[0].GetSpecificRuleId())
}

func TestMatch_BRPastTenseSetsSpecificRuleId(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(PortugueseBRDict, 1)
	sp.AddWord("casa")
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("falámos")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Equal(t, "MORFOLOGIK_RULE_PT_BR_DIALECT", m[0].GetSpecificRuleId())
}

func TestWirePortugueseSpellerTagger_Clitic(t *testing.T) {
	// Inject dict form with Java-style combined POS V…:P…
	wt := tagging.MapWordTagger{
		"diz-se": {tagging.NewTaggedWord("dizer", "VMIP3S0:PP3CSO00")},
	}
	tg := taggingpt.NewPortugueseTagger(wt)
	sp := morfologik.NewMorfologikSpeller(PortuguesePTDict, 1)
	// speller does NOT know diz-se → would flag without clitic check
	sp.AddWord("casa")
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	WirePortugueseSpellerTagger(r, tg)

	require.True(t, r.isValidCliticVerb("diz-se"))
	sent := languagetool.AnalyzePlain("diz-se")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m, "valid clitic verb must not be flagged")
}

func TestWirePortugueseSpellerTagger_CliticDialectLemmaInvalid(t *testing.T) {
	// lemma in dialect map → not a *valid* clitic for this dialect (Java)
	wt := tagging.MapWordTagger{
		"detetava-se": {tagging.NewTaggedWord("detetar", "VMII3S0:PP3CSO00")},
	}
	tg := taggingpt.NewPortugueseTagger(wt)
	r := NewMorfologikBrazilianPortugueseSpellerRule()
	WirePortugueseSpellerTagger(r, tg)
	// if "detetar" is a key in BR dialect map (European form), clitic is invalid
	if _, ok := r.dialectMap["detetar"]; !ok {
		// inject synthetic mapping for test
		if r.dialectMap == nil {
			r.dialectMap = map[string]string{}
		}
		r.dialectMap["detetar"] = "detectar"
	}
	require.False(t, r.isValidCliticVerb("detetava-se"))
}

func TestWordIsMisspelled_UsesRuleHook(t *testing.T) {
	r := NewMorfologikPortugalPortugueseSpellerRule()
	// empty map Speller would call every unknown misspelled
	r.IsMisspelled = func(w string) bool { return w == "xyzzy" }
	require.True(t, r.wordIsMisspelled("xyzzy"))
	require.False(t, r.wordIsMisspelled("casa"))
}

func TestWordSuggestions_MapThenFilter(t *testing.T) {
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.Speller.AddWord("casa")
	r.Speller.Suggestions["caza"] = []string{"casa"}
	require.Equal(t, []string{"casa"}, r.wordSuggestions("caza"))
	// without map suggestions and without filter dict → nil
	r2 := NewMorfologikPortugalPortugueseSpellerRule()
	require.Empty(t, r2.wordSuggestions("caza"))
}

func TestMatch_TitlecaseUsesRuleIsMisspelled(t *testing.T) {
	// lower form accepted via IsMisspelled hook (FilterDict-style), map Speller empty
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.IsMisspelled = func(w string) bool {
		// only lower guarda-chuva is known
		return strings.ToLower(w) != "guarda-chuva"
	}
	sent := languagetool.AnalyzePlain("Guarda-Chuva")
	// without hook path: empty Words means IsMisspelled on Speller true for all → would flag
	// with IsMisspelled: lower accepted → drop
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestCheckCompoundElements_UsesWordSuggestions(t *testing.T) {
	r := NewMorfologikPortugalPortugueseSpellerRule()
	r.IsMisspelled = func(w string) bool { return w == "chuvaa" }
	r.Speller.Suggestions["chuvaa"] = []string{"chuva"}
	// guarda known (not misspelled), chuvaa → chuva
	got := r.checkCompoundElements([]string{"guarda", "chuvaa"})
	require.Equal(t, "guarda-chuva", got)
}
