// Package corepack registers language-specific core rule packs for Check.
// Lives outside language packages and base rules to avoid import cycles while
// letting CLI and server share one dispatch table.
package corepack

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ar"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ca"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/da"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/el"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/es"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/fr"
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
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/sv"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
)

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
	default:
		rules.RegisterCoreRules(lt, lang)
	}
}
