package identifier

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier/detector"
	"github.com/stretchr/testify/require"
)

func TestDefaultLanguageIdentifier(t *testing.T) {
	// canLanguageBeDetected needs registry entries
	for _, c := range []string{"en", "de", "fr"} {
		if !languagetool.GlobalLanguages.IsLanguageSupported(c) {
			languagetool.GlobalLanguages.Register(languagetool.LanguageMeta{Name: c, Code: c})
		}
	}
	d := NewDefaultLanguageIdentifier(1000)
	d.ProfileScore = func(text string, preferred []string) map[string]float64 {
		if containsStr(preferred, "de") && javaStringLen(text) <= ConsiderOnlyPreferredThreshold {
			return map[string]float64{"de": 0.9}
		}
		return map[string]float64{"en": 0.95, "de": 0.1}
	}
	got := d.DetectLanguage("Hello world this is English text for detection.", nil, nil)
	require.NotNil(t, got)
	require.Equal(t, "en", got.DetectedLanguageCode)

	// short + preferred
	got2 := d.DetectLanguage("Hi", nil, []string{"de"})
	require.NotNil(t, got2)
	require.Equal(t, "de", got2.DetectedLanguageCode)

	// with ngram (char fallback under NGramDetector)
	ng := detector.NewNGramDetector(500)
	ng.TrainFromText("en", "hello world hello world hello")
	ng.TrainFromText("fr", "bonjour le monde bonjour le monde")
	d2 := NewDefaultLanguageIdentifier(500)
	d2.NGram = ng
	got3 := d2.Detect(d2.CleanAndShortenText("hello world hello"), nil, nil)
	require.NotNil(t, got3)
	require.Equal(t, "en", got3.DetectedLanguageCode)
}

func TestDefaultLanguageIdentifier_PrepareDetectUnsupported(t *testing.T) {
	// Java prepareDetectLanguage: dominant "th" → null → NoopLanguage (zz)
	d := NewDefaultLanguageIdentifier(1000)
	d.Unicode = detector.NewUnicodeBasedDetector()
	// Thai-heavy text
	thai := "ภาษาไทยภาษาไทยภาษาไทยภาษาไทยภาษาไทย"
	scores := d.Scores(thai, []string{}, []string{"en"}, false, 3)
	require.NotEmpty(t, scores)
	require.Equal(t, "zz", scores[0].DetectedLanguageCode)
}

// Ports detectLanguageCode + textObjectFactory fallback when no fasttext/ngram.
func TestDefaultLanguageIdentifier_DetectLanguageCodeFallback(t *testing.T) {
	for _, c := range []string{"en", "de"} {
		if !languagetool.GlobalLanguages.IsLanguageSupported(c) {
			languagetool.GlobalLanguages.Register(languagetool.LanguageMeta{Name: c, Code: c})
		}
	}
	d := NewDefaultLanguageIdentifier(1000)
	// No fastText / ngram → always +fallback + detectLanguageCode (ProfileScore)
	d.ProfileScore = func(text string, preferred []string) map[string]float64 {
		// Minority Cyrillic mixed into Latin should be stripped by textObjectFactory
		// before ProfileScore sees text (Latin-dominant).
		if strings.Contains(text, "привет") {
			return map[string]float64{"ru": 0.99}
		}
		return map[string]float64{"en": 0.9, "de": 0.1}
	}
	got := d.Detect("This is clearly English text for detection purposes.", nil, nil)
	require.NotNil(t, got)
	require.Equal(t, "en", got.DetectedLanguageCode)
	require.NotNil(t, got.GetDetectionSource())
	require.Contains(t, *got.GetDetectionSource(), "fallback")

	// limitOnPreferredLangs filters optimaize list to preferred only
	got2 := d.DetectLimit("This is English.", nil, []string{"de"}, true)
	require.NotNil(t, got2)
	require.Equal(t, "de", got2.DetectedLanguageCode)

	// Mixed Latin + minority Cyrillic: minority scripts removed → not russian
	mixed := "Hello world " + strings.Repeat("x", 20) + " привет"
	got3 := d.Detect(mixed, nil, nil)
	require.NotNil(t, got3)
	require.Equal(t, "en", got3.DetectedLanguageCode)
}

func TestApplyTextObjectFactoryFilters_MinorityScripts(t *testing.T) {
	// Mostly Latin with a few Cyrillic letters → Cyrillic minority removed
	in := "Hello world and some latin text а" // one Cyrillic letter
	out := ApplyTextObjectFactoryFilters(in)
	require.NotContains(t, out, "а")
	require.Contains(t, out, "Hello")
}

// Ports reinitFasttextAfterFailure when RunFasttext returns FastTextException(disabled).
func TestDefaultLanguageIdentifier_FastTextFailureFallback(t *testing.T) {
	for _, c := range []string{"en", "de"} {
		if !languagetool.GlobalLanguages.IsLanguageSupported(c) {
			languagetool.GlobalLanguages.Register(languagetool.LanguageMeta{Name: c, Code: c})
		}
	}
	ft := detector.NewFastTextDetectorForTest()
	calls := 0
	ft.Runner = func(line string) (string, error) {
		calls++
		return "", detector.NewFastTextException("disabled", true)
	}
	d := NewDefaultLanguageIdentifier(1000)
	d.SetFastTextDetector(ft)
	d.ProfileScore = func(text string, preferred []string) map[string]float64 {
		return map[string]float64{"en": 0.9}
	}
	long := strings.Repeat("hello world ", 20)
	got := d.Detect(long, nil, nil)
	require.NotNil(t, got)
	require.Equal(t, "en", got.DetectedLanguageCode)
	// Fallback path used after fasttext failure
	require.NotNil(t, got.GetDetectionSource())
	require.Contains(t, *got.GetDetectionSource(), "fallback")
	require.GreaterOrEqual(t, calls, 1)
}

// Ports DefaultLanguageIdentifier: runFasttext(text, additionalLangs) + fasttext conf rewrite.
func TestDefaultLanguageIdentifier_FastTextAdditionalAndConf(t *testing.T) {
	for _, c := range []string{"en", "de", "fr"} {
		if !languagetool.GlobalLanguages.IsLanguageSupported(c) {
			languagetool.GlobalLanguages.Register(languagetool.LanguageMeta{Name: c, Code: c})
		}
	}
	ft := detector.NewFastTextDetectorForTest()
	ft.Runner = func(line string) (string, error) {
		return "__label__en 0.99 __label__de 0.01\n", nil
	}
	d := NewDefaultLanguageIdentifier(1000)
	d.SetFastTextDetector(ft)
	// long text → fasttext path
	long := strings.Repeat("hello world ", 20)
	got := d.Detect(long, nil, nil)
	require.NotNil(t, got)
	require.Equal(t, "en", got.DetectedLanguageCode)
	// Java: 0.99 / (30.0 / min(len, 30)) for fasttext single result
	// text length UTF-16 of long > 30 → 0.99 / (30/30) = 0.99
	require.InDelta(t, 0.99, float64(got.GetDetectionConfidence()), 1e-5)
	require.NotNil(t, got.GetDetectionSource())
	require.Contains(t, *got.GetDetectionSource(), "fasttext")

	// noop additional "xx" alone would not register en drop; additional zz kept only if listed
	ft.Runner = func(line string) (string, error) {
		return "__label__en 0.5 __label__zz 0.4\n", nil
	}
	got2 := d.Detect(long, []string{"zz"}, nil)
	require.NotNil(t, got2)
	// en still preferred as higher score among canDetect
	require.Equal(t, "en", got2.DetectedLanguageCode)
}

func TestDefaultLanguageIdentifier_NoInventHasNonLatin(t *testing.T) {
	// Invent hasNonLatin score boost removed — Cyrillic still detected via PrepareDetect
	// expanding preferred/additional, not invent 0.6 score injection.
	d := NewDefaultLanguageIdentifier(1000)
	d.ProfileScore = func(text string, preferred []string) map[string]float64 {
		// Prefer list may include "ru" from dominant codes
		out := map[string]float64{"en": 0.5}
		if containsStr(preferred, "ru") {
			out["ru"] = 0.95
		}
		return out
	}
	got := d.Detect("Привет мир это русский текст для теста", nil, []string{"en"})
	require.NotNil(t, got)
	// With preferred en only originally, PrepareDetect adds ru from unicode dominant
	// so profile can score ru highly.
	require.Equal(t, "ru", got.DetectedLanguageCode)
}
