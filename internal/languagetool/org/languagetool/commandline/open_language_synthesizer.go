package commandline

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	arsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ar"
	casynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/ca"
	desynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/de"
	ensynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/en"
	essynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/es"
	frsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/fr"
	plsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pl"
)

// OpenLanguageSynthesizer ports Language.createDefaultSynthesizer resource open:
// language-specific synthesizer type when Java uses one (EN/DE/PL getPosTagCorrection /
// compound forms); otherwise BaseSynthesizer over the official *_synth.dict.
// Returns nil when dictPath is empty or the binary cannot be opened (fail-closed).
func OpenLanguageSynthesizer(langShort, dictPath string) synthesis.Synthesizer {
	if dictPath == "" {
		return nil
	}
	base := strings.ToLower(langShort)
	if i := strings.IndexByte(base, '-'); i > 0 {
		base = base[:i]
	}
	switch base {
	case "en":
		if s := ensynth.OpenEnglishSynthesizerFromDictPath(dictPath); s != nil {
			return s
		}
		return nil
	case "de":
		if s := desynth.OpenGermanSynthesizerFromDictPath(dictPath); s != nil {
			return s
		}
		return nil
	case "pl":
		// PolishSynthesizer.getPosTagCorrection for setpos (not plain BaseSynthesizer).
		if s := plsynth.OpenPolishSynthesizerFromDictPath(dictPath); s != nil {
			return s
		}
		return nil
	case "ar":
		// ArabicSynthesizer.getPosTagCorrection → correctTag (conj/def/pronoun flags).
		if s := arsynth.OpenArabicSynthesizerFromDictPath(dictPath); s != nil {
			return s
		}
		return nil
	case "fr":
		// FrenchSynthesizer.isException filter (qq*, trailing è).
		if s := frsynth.OpenFrenchSynthesizerFromDictPath(dictPath); s != nil {
			return s
		}
		return nil
	case "es":
		// SpanishSynthesizer: verb lemma "verb rest" + getTargetPosTag comparator.
		if s := essynth.OpenSpanishSynthesizerFromDictPath(dictPath); s != nil {
			return s
		}
		return nil
	case "ca":
		if s := casynth.OpenCatalanSynthesizerFromDictPath(dictPath); s != nil {
			return s
		}
		return nil
	default:
		if s := synthesis.OpenBaseSynthesizerFromDictPath(base, dictPath); s != nil {
			return s
		}
		return nil
	}
}
