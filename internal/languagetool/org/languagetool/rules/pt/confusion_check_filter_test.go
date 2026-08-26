package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func atrWithPOS(surface, pos string) *languagetool.AnalyzedTokenReadings {
	p := pos
	return languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken(surface, &p, nil))
}

func confusionMatch(langCode, message, short, template string) *rules.RuleMatch {
	pr := patterns.NewPatternRule("CONFUSION_TEST", langCode, nil, "desc", message, short)
	m := rules.NewRuleMatch(pr, nil, 0, 5, message)
	m.SetShortMessage(short)
	m.SetType(rules.RuleMatchTypeOther)
	m.SetSuggestedReplacements([]string{template})
	return m
}

func TestConfusionCheckFilter_LoadsBaseAndVariantMaps(t *testing.T) {
	f := NewConfusionCheckFilter()
	require.NotNil(t, f.relevantWordsPT)
	require.NotNil(t, f.relevantWordsBR)

	// Base key present in both maps (from /pt/confusion_pairs.txt).
	require.Contains(t, f.relevantWordsPT, "abobora")
	require.Contains(t, f.relevantWordsBR, "abobora")
	require.Equal(t, "abóbora", f.relevantWordsPT["abobora"].GetToken())
	require.Equal(t, "abóbora", f.relevantWordsBR["abobora"].GetToken())

	// Variant-only key "anonimo": orthography differs BR vs PT (not in base).
	require.Contains(t, f.relevantWordsPT, "anonimo")
	require.Contains(t, f.relevantWordsBR, "anonimo")
	require.Equal(t, "anónimo", f.relevantWordsPT["anonimo"].GetToken())
	require.Equal(t, "anônimo", f.relevantWordsBR["anonimo"].GetToken())

	// Routing selects distinct orthographies for the same key (BR vs PT maps).
	require.NotEqual(t, f.relevantWordsPT["anonimo"].GetToken(), f.relevantWordsBR["anonimo"].GetToken())
}

func TestConfusionCheckFilter_VariantRouting(t *testing.T) {
	f := NewConfusionCheckFilter()
	args := map[string]string{"postag": "N.*|A.*", "form": "anonimo"}

	// pt-BR → BR map → anônimo
	outBR := f.AcceptRuleMatch(confusionMatch("pt-BR", "msg se escribe con tilde", "short", "{suggestion}"),
		args, 0, nil, nil)
	require.NotNil(t, outBR)
	require.Equal(t, []string{"anônimo"}, outBR.GetSuggestedReplacements())

	// pt-PT → PT map → anónimo
	outPT := f.AcceptRuleMatch(confusionMatch("pt-PT", "msg se escribe con tilde", "short", "{suggestion}"),
		args, 0, nil, nil)
	require.NotNil(t, outPT)
	require.Equal(t, []string{"anónimo"}, outPT.GetSuggestedReplacements())

	// other (Africa etc.) → PT map (Java comment: PT spelling dominant)
	outOther := f.AcceptRuleMatch(confusionMatch("pt", "msg se escribe con tilde", "short", "{suggestion}"),
		args, 0, nil, nil)
	require.NotNil(t, outOther)
	require.Equal(t, []string{"anónimo"}, outOther.GetSuggestedReplacements())

	outAO := f.AcceptRuleMatch(confusionMatch("pt-AO", "msg se escribe con tilde", "short", "{suggestion}"),
		args, 0, nil, nil)
	require.NotNil(t, outAO)
	require.Equal(t, []string{"anónimo"}, outAO.GetSuggestedReplacements())
}

func TestConfusionCheckFilter_AcceptKeepVsNull(t *testing.T) {
	f := NewConfusionCheckFilter()

	// Keep: known form + matching postag, no gendernumberFrom.
	out := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "use tilde se escribe con tilde", "s", "{suggestion}"),
		map[string]string{"postag": "NC.*", "form": "abobora"},
		0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"abóbora"}, out.GetSuggestedReplacements())
	require.Equal(t, "s", out.GetShortMessage())
	require.Equal(t, rules.RuleMatchTypeOther, out.GetType())

	// Null: missing form
	require.Nil(t, f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m", "s", "{suggestion}"),
		map[string]string{"postag": "NC.*", "form": "zzz_not_a_confusion_key"},
		0, nil, nil))

	// Null: form present but postag does not match any reading
	require.Nil(t, f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m", "s", "{suggestion}"),
		map[string]string{"postag": "V.+", "form": "abobora"},
		0, nil, nil))
}

func TestConfusionCheckFilter_DiacriticMessageRewrite(t *testing.T) {
	// Official pairs always add diacritics to unaccented keys → keep "se escribe con tilde".
	f := NewConfusionCheckFilter()
	msgKeep := "La palabra se escribe con tilde."
	outKeep := f.AcceptRuleMatch(
		confusionMatch("pt-PT", msgKeep, "s", "{suggestion}"),
		map[string]string{"postag": "NC.*", "form": "abobora"},
		0, nil, nil)
	require.NotNil(t, outKeep)
	require.Equal(t, msgKeep, outKeep.GetMessage())
	require.Contains(t, outKeep.GetMessage(), "se escribe con tilde")

	// Rewrite branch: !(hasDiacritics(repl) && !hasDiacritics(form)).
	// Official data has no non-diacritic replacements; exercise AcceptRuleMatch
	// control flow with a minimal in-memory map twin of Java fields (test-only).
	pos := "NCMS000"
	atr := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("casa", &pos, nil))
	synth := &ConfusionCheckFilter{
		relevantWordsPT: map[string]*languagetool.AnalyzedTokenReadings{"casa": atr},
		relevantWordsBR: map[string]*languagetool.AnalyzedTokenReadings{},
	}
	msg := "La palabra se escribe con tilde."
	outRewrite := synth.AcceptRuleMatch(
		confusionMatch("pt-PT", msg, "s", "{suggestion}"),
		map[string]string{"postag": "NC.*", "form": "casa"},
		0, nil, nil)
	require.NotNil(t, outRewrite)
	require.Equal(t, "La palabra se escribe de otra manera.", outRewrite.GetMessage())
	require.Equal(t, []string{"casa"}, outRewrite.GetSuggestedReplacements())
}

func TestConfusionCheckFilter_SuggestionTemplates(t *testing.T) {
	f := NewConfusionCheckFilter()
	args := map[string]string{"postag": "NC.*", "form": "abobora"}

	out1 := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m se escribe con tilde", "s", "use {suggestion} please"),
		args, 0, nil, nil)
	require.NotNil(t, out1)
	require.Equal(t, []string{"use abóbora please"}, out1.GetSuggestedReplacements())

	out2 := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m se escribe con tilde", "s", "{Suggestion}"),
		args, 0, nil, nil)
	require.NotNil(t, out2)
	require.Equal(t, []string{"Abóbora"}, out2.GetSuggestedReplacements())

	out3 := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m se escribe con tilde", "s", "{SUGGESTION}"),
		args, 0, nil, nil)
	require.NotNil(t, out3)
	require.Equal(t, []string{"ABÓBORA"}, out3.GetSuggestedReplacements())

	// Combined templates
	out4 := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m se escribe con tilde", "s", "{suggestion}/{Suggestion}/{SUGGESTION}"),
		args, 0, nil, nil)
	require.NotNil(t, out4)
	require.Equal(t, []string{"abóbora/Abóbora/ABÓBORA"}, out4.GetSuggestedReplacements())
}

func TestConfusionCheckFilter_GenderNumberMatchAndMismatch(t *testing.T) {
	f := NewConfusionCheckFilter()
	// abobora → abóbora NCFS000; first reading POS is NCFS000
	// MS probe on pattern token → desired MS pattern does not match NCFS000 → null
	msTok := atrWithPOS("o", "DA0MS0") // [NAPD].+MS.*
	outMS := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m se escribe con tilde", "s", "{suggestion}"),
		map[string]string{"postag": "NC.*", "form": "abobora", "gendernumberFrom": "1"},
		0, []*languagetool.AnalyzedTokenReadings{msTok}, nil)
	require.Nil(t, outMS)

	// FS probe → desired FS matches NCFS000 → keep abóbora
	fsTok := atrWithPOS("a", "DA0FS0")
	outFS := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m se escribe con tilde", "s", "{suggestion}"),
		map[string]string{"postag": "NC.*", "form": "abobora", "gendernumberFrom": "1"},
		0, []*languagetool.AnalyzedTokenReadings{fsTok}, nil)
	require.NotNil(t, outFS)
	require.Equal(t, []string{"abóbora"}, outFS.GetSuggestedReplacements())
}

func TestConfusionCheckFilter_GenderNumberFromSetButNoBranch(t *testing.T) {
	// Java quirk: gendernumberFrom set, token POS matches none of the gender
	// branches → desiredGenderNumberPattern stays null → replacement stays null
	// even when form maps and postag matches.
	f := NewConfusionCheckFilter()
	// Verb present not participle V.P..* — fails all [NAPD]/ V.P probes
	vTok := atrWithPOS("corre", "VMIP3S0")
	out := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m se escribe con tilde", "s", "{suggestion}"),
		map[string]string{"postag": "NC.*", "form": "abobora", "gendernumberFrom": "1"},
		0, []*languagetool.AnalyzedTokenReadings{vTok}, nil)
	require.Nil(t, out)
}

func TestConfusionCheckFilter_GenderNumberBounds(t *testing.T) {
	f := NewConfusionCheckFilter()
	tok := atrWithPOS("a", "DA0FS0")
	args := map[string]string{"postag": "NC.*", "form": "abobora", "gendernumberFrom": "0"}
	require.Panics(t, func() {
		f.AcceptRuleMatch(
			confusionMatch("pt-PT", "m", "s", "{suggestion}"),
			args, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	})
	args2 := map[string]string{"postag": "NC.*", "form": "abobora", "gendernumberFrom": "2"}
	require.Panics(t, func() {
		f.AcceptRuleMatch(
			confusionMatch("pt-PT", "m", "s", "{suggestion}"),
			args2, 0, []*languagetool.AnalyzedTokenReadings{tok}, nil)
	})
}

func TestConfusionCheckFilter_FormLowercased(t *testing.T) {
	f := NewConfusionCheckFilter()
	out := f.AcceptRuleMatch(
		confusionMatch("pt-PT", "m se escribe con tilde", "s", "{suggestion}"),
		map[string]string{"postag": "NC.*", "form": "AbObOrA"},
		0, nil, nil)
	require.NotNil(t, out)
	require.Equal(t, []string{"abóbora"}, out.GetSuggestedReplacements())
}

func TestConfusionCheckFilter_MultiReadingFirstPOSOnly(t *testing.T) {
	// Java: desiredGenderNumberPattern.matcher(readings.get(0).getPOSTag()).matches()
	// anonimo BR first reading is AQ0MS0 (anônimo); second is NCMS000.
	// MS desired matches first POS → accept; FS desired fails first POS → null.
	f := NewConfusionCheckFilter()
	msTok := atrWithPOS("um", "DI0MS0")
	outMS := f.AcceptRuleMatch(
		confusionMatch("pt-BR", "m se escribe con tilde", "s", "{suggestion}"),
		map[string]string{"postag": "N.*|A.*", "form": "anonimo", "gendernumberFrom": "1"},
		0, []*languagetool.AnalyzedTokenReadings{msTok}, nil)
	require.NotNil(t, outMS)
	require.Equal(t, []string{"anônimo"}, outMS.GetSuggestedReplacements())

	fsTok := atrWithPOS("uma", "DI0FS0")
	outFS := f.AcceptRuleMatch(
		confusionMatch("pt-BR", "m se escribe con tilde", "s", "{suggestion}"),
		map[string]string{"postag": "N.*|A.*", "form": "anonimo", "gendernumberFrom": "1"},
		0, []*languagetool.AnalyzedTokenReadings{fsTok}, nil)
	require.Nil(t, outFS)
}

func TestConfusionCheckFilter_RegisterImplementsRuleFilter(t *testing.T) {
	var _ patterns.RuleFilter = NewConfusionCheckFilter()
	created := patterns.GlobalRuleFilterCreator.GetFilter("org.languagetool.rules.pt.ConfusionCheckFilter")
	require.NotNil(t, created)
	_, ok := created.(*ConfusionCheckFilter)
	require.True(t, ok)
}
