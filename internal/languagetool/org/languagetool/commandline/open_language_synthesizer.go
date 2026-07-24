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
	nlsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/nl"
	plsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pl"
	ptsynth "github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis/pt"
)

// OpenLanguageSynthesizer ports Language.createDefaultSynthesizer resource open:
// language-specific synthesizer type when Java uses one (EN/DE/PL getPosTagCorrection /
// compound forms, CA/ES verb lemma space, FR isException); otherwise BaseSynthesizer
// over the official *_synth.dict.
// langShort may be a full code (ca-ES-valencia) for Catalan regional verb tags.
// Returns nil when dictPath is empty or the binary cannot be opened (fail-closed).
func OpenLanguageSynthesizer(langShort, dictPath string) synthesis.Synthesizer {
	if dictPath == "" {
		return nil
	}
	full := langShort
	base := strings.ToLower(langShort)
	// Keep full code for Catalan variants (ca-ES-valencia); switch on first segment.
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
		// CatalanSynthesizer: LemmasToIgnore, regional verb tags, getTargetPosTag.
		if s := casynth.OpenCatalanSynthesizerFromDictPath(dictPath, full); s != nil {
			return s
		}
		return nil
	case "pt":
		if s := ptsynth.OpenPortugueseSynthesizerFromDictPath(dictPath); s != nil {
			return s
		}
		return nil
	case "nl":
		if s := nlsynth.OpenDutchSynthesizerFromDictPath(dictPath); s != nil {
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
