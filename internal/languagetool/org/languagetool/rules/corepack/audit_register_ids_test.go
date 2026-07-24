package corepack

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ast"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/be"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/br"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/crh"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/da"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/el"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/eo"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/es"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fa"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ga"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/gl"
	islang "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/is"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/it"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/km"
	ltlang "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/lt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ml"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/nl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ro"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ru"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sv"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/tl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/stretchr/testify/require"
)

// uniqueIDs collapses duplicate Java asList entries (e.g. Irish registers
// UppercaseSentenceStartRule twice with the same getId).
func uniqueIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// TestRegisterCore_MatchesRelevantRuleIDs locks RegisterCore* IDs to
// language.*RelevantRuleIDs / GetRelevantRuleIDs (Java getRelevantRules class getId).
// EN/DE also register createDefaultSpellingRule / LanguageModelCapable spellers
// (not in getRelevantRules asList) — those are the only allowed extras.
func TestRegisterCore_MatchesRelevantRuleIDs(t *testing.T) {
	type cT struct {
		name string
		code string
		reg  func(*languagetool.JLanguageTool)
		want []string
	}
	exact := []cT{
		{"ar", "ar", ar.RegisterCoreArabicRules, language.ArabicRelevantRuleIDs()},
		{"ast", "ast", ast.RegisterCoreAsturianRules, language.AsturianRelevantRuleIDs()},
		{"be", "be", be.RegisterCoreBelarusianRules, language.BelarusianRelevantRuleIDs()},
		{"br", "br", br.RegisterCoreBretonRules, language.BretonRelevantRuleIDs()},
		{"ca", "ca", ca.RegisterCoreCatalanRules, language.CatalanRelevantRuleIDs()},
		{"crh", "crh", crh.RegisterCoreCrimeanTatarRules, language.CrimeanTatarRelevantRuleIDs()},
		{"da", "da", da.RegisterCoreDanishRules, language.DanishRelevantRuleIDs()},
		{"el", "el", el.RegisterCoreGreekRules, language.GreekRelevantRuleIDs()},
		{"eo", "eo", eo.RegisterCoreEsperantoRules, language.EsperantoRelevantRuleIDs()},
		{"es", "es", es.RegisterCoreSpanishRules, language.SpanishRelevantRuleIDs()},
		{"fa", "fa", fa.RegisterCorePersianRules, language.PersianRelevantRuleIDs()},
		{"fr", "fr", fr.RegisterCoreFrenchRules, language.FrenchRelevantRuleIDs()},
		{"ga", "ga", ga.RegisterCoreIrishRules, language.IrishRelevantRuleIDs()},
		{"gl", "gl", gl.RegisterCoreGalicianRules, language.GalicianRelevantRuleIDs()},
		{"is", "is", islang.RegisterCoreIcelandicRules, language.IcelandicRelevantRuleIDs()},
		{"it", "it", it.RegisterCoreItalianRules, language.ItalianRelevantRuleIDs()},
		{"km", "km", km.RegisterCoreKhmerRules, language.KhmerRelevantRuleIDs()},
		{"lt", "lt", ltlang.RegisterCoreLithuanianRules, language.LithuanianRelevantRuleIDs()},
		{"ml", "ml", ml.RegisterCoreMalayalamRules, language.MalayalamRelevantRuleIDs()},
		{"nl", "nl", nl.RegisterCoreDutchRules, language.DutchRelevantRuleIDs()},
		{"pl", "pl", pl.RegisterCorePolishRules, language.PolishRelevantRuleIDs()},
		{"ro", "ro", ro.RegisterCoreRomanianRules, language.RomanianRelevantRuleIDs()},
		{"ru", "ru", ru.RegisterCoreRussianRules, language.RussianRelevantRuleIDs()},
		{"sk", "sk", sk.RegisterCoreSlovakRules, language.SlovakRelevantRuleIDs()},
		{"sl", "sl", sl.RegisterCoreSlovenianRules, language.SlovenianRelevantRuleIDs()},
		{"sv", "sv", sv.RegisterCoreSwedishRules, language.SwedishRelevantRuleIDs()},
		{"tl", "tl", tl.RegisterCoreTagalogRules, language.TagalogRelevantRuleIDs()},
		{"uk", "uk", uk.RegisterCoreUkrainianRules, language.UkrainianRelevantRuleIDs()},
		{"pt-PT", "pt-PT", pt.RegisterCorePortugueseRules, language.PortugalPortuguese.GetRelevantRuleIDs()},
		{"sr", "sr", sr.RegisterCoreSerbianRules, language.DefaultSerbian.GetRelevantRuleIDs()},
	}
	for _, c := range exact {
		c := c
		t.Run(c.name, func(t *testing.T) {
			lt := languagetool.NewJLanguageTool(c.code)
			c.reg(lt)
			require.ElementsMatch(t, uniqueIDs(c.want), lt.GetAllRegisteredRuleIDs())
		})
	}

	t.Run("en-US", func(t *testing.T) {
		lt := languagetool.NewJLanguageTool("en-US")
		en.RegisterCoreEnglishLanguageRules(lt)
		want := append([]string(nil), language.AmericanEnglish.GetRelevantRuleIDs()...)
		// Java createDefaultSpellingRule / getRelevantLanguageModelCapableRules speller.
		want = append(want, language.AmericanEnglish.SpellerRuleID)
		require.ElementsMatch(t, uniqueIDs(want), lt.GetAllRegisteredRuleIDs())
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "WHITESPACE_PUNCTUATION")
	})

	t.Run("de-DE", func(t *testing.T) {
		// Package tests opt out of upstream grammar; honor env if set.
		lt := languagetool.NewJLanguageTool("de-DE")
		de.RegisterCoreGermanRules(lt)
		want := append([]string(nil), language.GermanyGerman.GetRelevantRuleIDs()...)
		// Java createDefaultSpellingRule / getRelevantLanguageModelCapableRules speller.
		want = append(want, language.GermanyGerman.SpellerRuleID)
		// When UseUpstreamGrammar loads XML, extra pattern IDs appear — only
		// assert class-ID surface when upstream grammar is off.
		if languagetool.UseUpstreamGrammar() {
			ids := lt.GetAllRegisteredRuleIDs()
			for _, id := range uniqueIDs(want) {
				require.Contains(t, ids, id, "missing %s", id)
			}
			require.NotContains(t, ids, "WHITESPACE_PUNCTUATION")
			require.NotContains(t, ids, "DE_PROHIBITED_COMPOUNDS")
			require.NotContains(t, ids, "DE_CONFUSION_RULE")
			require.NotContains(t, ids, "DE_UPPER_CASE_NGRAM")
			return
		}
		require.ElementsMatch(t, uniqueIDs(want), lt.GetAllRegisteredRuleIDs())
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "WHITESPACE_PUNCTUATION")
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "SENTENCE_WHITESPACE")
		require.NotContains(t, lt.GetAllRegisteredRuleIDs(), "DOUBLE_PUNCTUATION")
	})

	t.Run("ja-zh-ta", func(t *testing.T) {
		for _, code := range []string{"ja", "zh"} {
			lt := languagetool.NewJLanguageTool(code)
			Register(lt, code)
			// Japanese and Chinese share the same two-ID getRelevantRules list.
			require.ElementsMatch(t, language.JapaneseRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
			require.ElementsMatch(t, language.ChineseRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
		}
		lt := languagetool.NewJLanguageTool("ta")
		Register(lt, "ta")
		require.ElementsMatch(t, language.TamilRelevantRuleIDs(), lt.GetAllRegisteredRuleIDs())
	})
}
