// Package corepack registers language-specific core rule packs for Check.
// Lives outside language packages and base rules to avoid import cycles while
// letting CLI and server share one dispatch table.
package corepack

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	// Side-effect: language init wires FilterFrenchRuleMatchesHook (French AI_FR_GGEC).
	_ "github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/eo" // DateCheckFilter init + RegisterCore
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
	// named packs with faithful spellers where Java registers one
	{"eo", "Esperanto"},
	{"is", "Icelandic"},
	{"lt", "Lithuanian"},
	{"ml", "Malayalam"},
	{"tl", "Tagalog"},
	{"ast", "Asturian"},
	{"crh", "Crimean Tatar"},
	// Java getRelevantRules is layout-only for these (no invent word-repeat packs)
	{"ja", "Japanese"},
	{"ta", "Tamil"},
	{"zh", "Chinese"},
}

// registerJapaneseChineseRelevant ports Japanese/Chinese.getRelevantRules:
// DoublePunctuationRule + MultipleWhitespaceRule only — no invent word-repeat /
// shared full layout that Java does not register.
func registerJapaneseChineseRelevant(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))
	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))
}

// registerTamilRelevant ports Tamil.getRelevantRules:
// CommaWhitespace, DoublePunctuation, MultipleWhitespace, LongSentence(50), SentenceWhitespace.
func registerTamilRelevant(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	cw := rules.NewCommaWhitespaceRule(nil)
	lt.AddRuleChecker(cw.GetID(), rules.AsSentenceCheckerSimple(cw.Match))
	dp := rules.NewDoublePunctuationRule(nil)
	lt.AddRuleChecker(dp.GetID(), rules.AsSentenceCheckerSimple(dp.Match))
	ws := rules.NewMultipleWhitespaceRule(map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
	})
	lt.AddRuleChecker(ws.GetID(), rules.AsSentenceCheckerSimple(func(s *languagetool.AnalyzedSentence) []*rules.RuleMatch {
		return ws.Match([]*languagetool.AnalyzedSentence{s})
	}))
	// Java: new LongSentenceRule(messages, userConfig, 50) — text-level
	ls := rules.NewLongSentenceRule(nil, 50)
	lt.AddTextLevelRuleChecker(ls.GetID(), rules.AsTextLevelChecker(ls.MatchList))
	sw := rules.NewSentenceWhitespaceRule(nil)
	lt.AddTextLevelRuleChecker(sw.GetID(), rules.AsTextLevelChecker(sw.MatchList))
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
		eo.RegisterCoreEsperantoRules(lt)
	case "is":
		islang.RegisterCoreIcelandicRules(lt)
	case "lt":
		ltlang.RegisterCoreLithuanianRules(lt)
	case "ml":
		ml.RegisterCoreMalayalamRules(lt)
	case "tl":
		tl.RegisterCoreTagalogRules(lt)
	case "ast":
		ast.RegisterCoreAsturianRules(lt)
	case "crh":
		crh.RegisterCoreCrimeanTatarRules(lt)
	case "sr":
		sr.RegisterCoreSerbianRules(lt)
	case "ja", "zh":
		// Java Chinese/Japanese.getRelevantRules — no invent word-repeat
		registerJapaneseChineseRelevant(lt)
	case "ta":
		registerTamilRelevant(lt)
	default:
		// No invent SharedLayout / WordRepeat for unknown languages.
		// Java only registers rules listed in that language's getRelevantRules.
		_ = lang
	}
}
