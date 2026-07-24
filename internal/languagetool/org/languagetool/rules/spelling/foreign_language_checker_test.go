package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestForeignLanguageChecker(t *testing.T) {
	c := NewForeignLanguageChecker("en", "hello world there", 10, []string{"en", "de"})
	// low error ratio
	require.Empty(t, c.Check(1))
	// high error ratio, no detector
	require.Empty(t, c.Check(5))

	c.Detect = func(sentence string, preferred []string, max int) []DetectedLanguageScore {
		return []DetectedLanguageScore{{ShortCode: "de", Confidence: 0.9}}
	}
	m := c.Check(5)
	require.Equal(t, float32(0.9), m["de"])

	c.Detect = func(sentence string, preferred []string, max int) []DetectedLanguageScore {
		return []DetectedLanguageScore{{ShortCode: "en", Confidence: 0.95}}
	}
	m = c.Check(5)
	require.Equal(t, float32(0.99), m[NoForeignLangDetected])
}
