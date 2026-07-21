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

	// with ngram
	ng := detector.NewCharNGramDetector(3)
	ng.TrainFromText("en", "hello world hello world hello")
	ng.TrainFromText("fr", "bonjour le monde bonjour le monde")
	d2 := NewDefaultLanguageIdentifier(500)
	d2.NGram = ng
	got3 := d2.Detect(d2.CleanAndShortenText("hello world hello"), nil, nil)
	require.NotNil(t, got3)
	require.Equal(t, "en", got3.DetectedLanguageCode)
}
