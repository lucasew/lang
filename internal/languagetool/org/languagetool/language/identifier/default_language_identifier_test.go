package identifier

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language/identifier/detector"
	"github.com/stretchr/testify/require"
)

func TestDefaultLanguageIdentifier(t *testing.T) {
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
