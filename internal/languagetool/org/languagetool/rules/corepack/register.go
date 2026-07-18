// Package corepack registers language-specific core rule packs for Check.
// Lives outside language packages and base rules to avoid import cycles while
// letting CLI and server share one dispatch table.
package corepack

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/be"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/br"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/da"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/el"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/es"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fa"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ga"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/gl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/it"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/km"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/nl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/pt"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ro"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ru"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sk"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sl"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sr"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sv"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
)

// Supported lists short codes with dedicated core packs (for /v2/languages etc.).
var Supported = []struct {
	Code string
	Name string
}{
	{"en", "English"},
	{"de", "German"},
	{"fr", "French"},
	{"es", "Spanish"},
	{"nl", "Dutch"},
	{"pl", "Polish"},
	{"uk", "Ukrainian"},
	{"it", "Italian"},
	{"pt", "Portuguese"},
	{"ru", "Russian"},
	{"ca", "Catalan"},
	{"sv", "Swedish"},
	{"da", "Danish"},
	{"gl", "Galician"},
	{"sk", "Slovak"},
	{"el", "Greek"},
	{"ro", "Romanian"},
	{"ar", "Arabic"},
	{"km", "Khmer"},
	{"sl", "Slovenian"},
	{"br", "Breton"},
	{"fa", "Persian"},
	{"ga", "Irish"},
	{"sr", "Serbian"},
	{"be", "Belarusian"},
	// generic layout + word-repeat packs (no language-specific rule twins yet)
	{"eo", "Esperanto"},
	{"is", "Icelandic"},
	{"ja", "Japanese"},
	{"lt", "Lithuanian"},
	{"ml", "Malayalam"},
	{"ta", "Tamil"},
	{"tl", "Tagalog"},
	{"zh", "Chinese"},
	{"ast", "Asturian"},
	{"crh", "Crimean Tatar"},
}

// registerGeneric installs shared layout + base word-repeat + word-repeat-beginning.
func registerGeneric(lt *languagetool.JLanguageTool, lang, wordRepeatID string) {
	rules.RegisterSharedLayoutRules(lt, lang)
	wr := rules.NewWordRepeatRule(map[string]string{"repetition": "Word repetition"})
	if wordRepeatID != "" {
		wr.IDOverride = wordRepeatID
	}
	lt.AddRuleChecker(wr.GetID(), rules.AsSentenceCheckerSimple(wr.Match))
	// soft text-level beginning rule for languages without a dedicated pack
	wrb := rules.NewWordRepeatBeginningRule(map[string]string{
		"desc_repetition_beginning_word": "Three successive sentences begin with the same word.",
	})
	if wordRepeatID != "" {
		// e.g. BE_WORD_REPEAT_RULE → BE_WORD_REPEAT_BEGINNING_RULE
		base := wordRepeatID
		if len(base) > 5 && base[len(base)-5:] == "_RULE" {
			base = base[:len(base)-5]
		}
		wrb.IDOverride = base + "_BEGINNING_RULE"
	}
	lt.AddTextLevelRuleChecker(wrb.GetID(), rules.AsTextLevelChecker(wrb.MatchList))
}

// Register installs the best available core rule pack for lang (e.g. "en-US", "de").
func Register(lt *languagetool.JLanguageTool, lang string) {
	if lt == nil {
		return
	}
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	switch strings.ToLower(base) {
	case "en":
		en.RegisterCoreEnglishLanguageRules(lt)
	case "de":
		de.RegisterCoreGermanRules(lt)
	case "fr":
		fr.RegisterCoreFrenchRules(lt)
	case "es":
		es.RegisterCoreSpanishRules(lt)
	case "nl":
		nl.RegisterCoreDutchRules(lt)
	case "pl":
		pl.RegisterCorePolishRules(lt)
	case "uk":
		uk.RegisterCoreUkrainianRules(lt)
	case "it":
		it.RegisterCoreItalianRules(lt)
	case "pt":
		pt.RegisterCorePortugueseRules(lt)
	case "ru":
		ru.RegisterCoreRussianRules(lt)
	case "ca":
		ca.RegisterCoreCatalanRules(lt)
	case "sv":
		sv.RegisterCoreSwedishRules(lt)
	case "da":
		da.RegisterCoreDanishRules(lt)
	case "gl":
		gl.RegisterCoreGalicianRules(lt)
	case "sk":
		sk.RegisterCoreSlovakRules(lt)
	case "el":
		el.RegisterCoreGreekRules(lt)
	case "ro":
		ro.RegisterCoreRomanianRules(lt)
	case "ar":
		ar.RegisterCoreArabicRules(lt)
	case "km":
		km.RegisterCoreKhmerRules(lt)
	case "sl":
		sl.RegisterCoreSlovenianRules(lt)
	case "br":
		br.RegisterCoreBretonRules(lt)
	case "fa":
		fa.RegisterCorePersianRules(lt)
	case "ga":
		ga.RegisterCoreIrishRules(lt)
	case "be":
		be.RegisterCoreBelarusianRules(lt)
	case "eo":
		registerGeneric(lt, "eo", "EO_WORD_REPEAT_RULE")
	case "is":
		registerGeneric(lt, "is", "IS_WORD_REPEAT_RULE")
	case "ja":
		registerGeneric(lt, "ja", "JA_WORD_REPEAT_RULE")
	case "lt":
		registerGeneric(lt, "lt", "LT_WORD_REPEAT_RULE")
	case "ml":
		registerGeneric(lt, "ml", "ML_WORD_REPEAT_RULE")
	case "sr":
		sr.RegisterCoreSerbianRules(lt)
	case "ta":
		registerGeneric(lt, "ta", "TA_WORD_REPEAT_RULE")
	case "tl":
		registerGeneric(lt, "tl", "TL_WORD_REPEAT_RULE")
	case "zh":
		registerGeneric(lt, "zh", "ZH_WORD_REPEAT_RULE")
	case "ast":
		registerGeneric(lt, "ast", "AST_WORD_REPEAT_RULE")
	case "crh":
		registerGeneric(lt, "crh", "CRH_WORD_REPEAT_RULE")
	default:
		rules.RegisterCoreRules(lt, lang)
	}
}
